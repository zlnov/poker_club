package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"poker-club/backend/internal/domain"
	"poker-club/backend/internal/service"
)

// --- Phase 03: Game management handlers ---

// handleCreateGame starts the game creation flow by showing the interactive
// game creation screen with all parameters editable.
func (b *Bot) handleCreateGame(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, err := strconv.ParseInt(strings.TrimPrefix(cb.Data, cbCreateGame+":"), 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}

	// Verify the user has permission to create games.
	_, err = b.svc.GetUserRole(ctx, cb.From.ID, clubID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: вы не являетесь участником клуба.")
		return
	}

	// Initialize game creation data with defaults.
	data := &gameCreationData{
		gameType:       "cash",
		bankerName:     "не выбран",
		currency:       "RUB",
		moneyModel:     "real",
		chipValue:      "1",
		buyInAmount:    "",
		rebuyAllowed:   false,
		rebuyPrice:     "",
		maxRebuys:      "",
		duration:       "",
		startTime:      "",
		minPlayers:     "",
		maxPlayers:     "",
		rankingPrimary: "profit",
	}

	b.setState(cb.From.ID, stateCreateGame, clubID)
	b.states[cb.From.ID].gameData = data

	b.sendTextWithKeyboard(cb.Message.Chat.ID, "Создание игры:", gameCreationKeyboard(data, clubID, nil))
}

// handleGameEditParam processes clicks on parameter fields in the game creation screen.
func (b *Bot) handleGameEditParam(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_edit_param:<club_id>:<param>
	rest := strings.TrimPrefix(cb.Data, cbGameEditParam+":")
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	param := parts[1]

	state := b.getState(cb.From.ID)
	if state == nil || state.gameData == nil {
		b.sendText(cb.Message.Chat.ID, "Сессия создания игры не найдена. Начните заново.")
		return
	}

	switch param {
	case "type":
		b.sendTextWithKeyboard(cb.Message.Chat.ID, "Выберите тип игры:", gameTypeKeyboard(clubID))
	case "banker":
		members, err := b.svc.GetClubMembers(ctx, cb.From.ID, clubID)
		if err != nil {
			b.log.Error("failed to get club members", "error", err)
			b.sendText(cb.Message.Chat.ID, "Ошибка при получении участников клуба.")
			return
		}
		var activeMembers []*domain.ClubMemberWithPlayer
		for _, m := range members {
			if m.Status == "active" {
				activeMembers = append(activeMembers, m)
			}
		}
		if len(activeMembers) == 0 {
			b.sendText(cb.Message.Chat.ID, "В клубе нет активных участников.")
			return
		}
		b.sendTextWithKeyboard(cb.Message.Chat.ID, "Выберите банкира:", bankerSelectKeyboard(clubID, activeMembers))
	case "currency":
		b.sendTextWithKeyboard(cb.Message.Chat.ID, "Выберите валюту:", currencyKeyboard(clubID))
	case "money_model":
		b.sendTextWithKeyboard(cb.Message.Chat.ID, "Выберите денежную модель:", moneyModelKeyboard(clubID))
	case "rebuy_allowed":
		b.sendTextWithKeyboard(cb.Message.Chat.ID, "Разрешен rebuy?", rebuyAllowedKeyboard(clubID))
	case "chip_value", "buy_in", "rebuy_price", "max_rebuys", "start_time", "duration", "min_players", "max_players", "ranking_primary":
		state.gameData.editingParam = param
		b.sendText(cb.Message.Chat.ID, paramPrompt(param))
	}
}

// handleGameSelectParam processes inline keyboard selections during game creation.
func (b *Bot) handleGameSelectParam(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_select_param:<club_id>:<param>:<value>
	rest := strings.TrimPrefix(cb.Data, cbGameSelectParam+":")
	parts := strings.SplitN(rest, ":", 3)
	if len(parts) != 3 {
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return
	}
	param := parts[1]
	value := parts[2]

	state := b.getState(cb.From.ID)
	if state == nil || state.gameData == nil {
		return
	}

	switch param {
	case "type":
		state.gameData.gameType = value
	case "banker":
		bankerID, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return
		}
		state.gameData.bankerID = bankerID
		// Find the banker name from club members.
		members, err := b.svc.GetClubMembers(ctx, cb.From.ID, clubID)
		if err == nil {
			for _, m := range members {
				if m.ID == bankerID {
					state.gameData.bankerName = memberShortName(m.Player)
					break
				}
			}
		}
	case "currency":
		state.gameData.currency = value
	case "money_model":
		state.gameData.moneyModel = value
	case "rebuy_allowed":
		state.gameData.rebuyAllowed = (value == "yes")
	}

	// Update the message with the new keyboard.
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Создание игры:", gameCreationKeyboard(state.gameData, clubID, nil))
}

// handleGameParamInput processes text input for game parameters.
func (b *Bot) handleGameParamInput(ctx context.Context, msg *tgbotapi.Message, state *userState) {
	data := state.gameData
	if data == nil {
		b.sendText(msg.Chat.ID, "Сессия создания игры не найдена.")
		return
	}

	param := data.editingParam
	value := strings.TrimSpace(msg.Text)

	switch param {
	case "chip_value":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			b.sendText(msg.Chat.ID, "Ошибка: введите число.")
			return
		}
		data.chipValue = value
	case "buy_in":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			b.sendText(msg.Chat.ID, "Ошибка: введите число.")
			return
		}
		data.buyInAmount = value
	case "rebuy_price":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			b.sendText(msg.Chat.ID, "Ошибка: введите число.")
			return
		}
		data.rebuyPrice = value
	case "max_rebuys":
		if _, err := strconv.Atoi(value); err != nil {
			b.sendText(msg.Chat.ID, "Ошибка: введите целое число.")
			return
		}
		data.maxRebuys = value
	case "start_time":
		if _, err := time.Parse("02.01.2006 15:04", value); err != nil {
			b.sendText(msg.Chat.ID, "Ошибка: неверный формат. Используйте: 15.08.2026 19:00")
			return
		}
		data.startTime = value
	case "duration":
		if value == "-" {
			data.duration = ""
		} else {
			if _, err := parseDuration(value); err != nil {
				b.sendText(msg.Chat.ID, "Ошибка: неверный формат. Пример: 5 часов")
				return
			}
			data.duration = value
		}
	case "min_players":
		if _, err := strconv.Atoi(value); err != nil {
			b.sendText(msg.Chat.ID, "Ошибка: введите целое число.")
			return
		}
		data.minPlayers = value
	case "max_players":
		if _, err := strconv.Atoi(value); err != nil {
			b.sendText(msg.Chat.ID, "Ошибка: введите целое число.")
			return
		}
		data.maxPlayers = value
	case "ranking_primary":
		data.rankingPrimary = value
	default:
		b.sendText(msg.Chat.ID, "Неизвестный параметр.")
		return
	}

	data.editingParam = ""
	b.setState(msg.From.ID, stateCreateGame, state.clubID)
	b.states[msg.From.ID].gameData = data

	b.sendTextWithKeyboard(msg.Chat.ID, "Создание игры:", gameCreationKeyboard(data, state.clubID, nil))
}

// handleGameCreateConfirm processes the "Создать" button in the game creation screen.
func (b *Bot) handleGameCreateConfirm(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, err := strconv.ParseInt(strings.TrimPrefix(cb.Data, cbGameCreateConfirm+":"), 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}

	state := b.getState(cb.From.ID)
	if state == nil || state.gameData == nil {
		b.sendText(cb.Message.Chat.ID, "Сессия создания игры не найдена.")
		return
	}

	data := state.gameData

	// Validate required fields.
	if data.bankerID == 0 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: не выбран банкир.")
		return
	}
	if data.buyInAmount == "" {
		b.sendText(cb.Message.Chat.ID, "Ошибка: не указана стоимость buy-in.")
		return
	}
	if data.startTime == "" {
		b.sendText(cb.Message.Chat.ID, "Ошибка: не указано время начала.")
		return
	}
	if data.minPlayers == "" {
		b.sendText(cb.Message.Chat.ID, "Ошибка: не указано минимальное количество игроков.")
		return
	}
	if data.maxPlayers == "" {
		b.sendText(cb.Message.Chat.ID, "Ошибка: не указано максимальное количество игроков.")
		return
	}

	// Parse start time.
	startTime, err := time.Parse("02.01.2006 15:04", data.startTime)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный формат времени. Используйте: 15.08.2026 19:00")
		return
	}

	// Parse buy-in amount.
	buyInAmount, err := strconv.ParseFloat(data.buyInAmount, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверное значение buy-in.")
		return
	}

	// Parse chip value.
	chipValue, err := strconv.ParseFloat(data.chipValue, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверное значение chip value.")
		return
	}

	// Parse min/max players.
	minPlayers, err := strconv.Atoi(data.minPlayers)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверное значение min players.")
		return
	}
	maxPlayers, err := strconv.Atoi(data.maxPlayers)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверное значение max players.")
		return
	}

	// Parse duration (optional).
	var duration *time.Duration
	if data.duration != "" {
		d, err := parseDuration(data.duration)
		if err != nil {
			b.sendText(cb.Message.Chat.ID, "Ошибка: неверный формат продолжительности. Пример: 5 часов")
			return
		}
		duration = &d
	}

	// Parse rebuy price (optional).
	var rebuyPrice *float64
	if data.rebuyAllowed && data.rebuyPrice != "" {
		rp, err := strconv.ParseFloat(data.rebuyPrice, 64)
		if err != nil {
			b.sendText(cb.Message.Chat.ID, "Ошибка: неверное значение rebuy price.")
			return
		}
		rebuyPrice = &rp
	}

	// Parse max rebuys (optional).
	var maxRebuys *int
	if data.rebuyAllowed && data.maxRebuys != "" {
		mr, err := strconv.Atoi(data.maxRebuys)
		if err != nil {
			b.sendText(cb.Message.Chat.ID, "Ошибка: неверное значение max rebuys.")
			return
		}
		maxRebuys = &mr
	}

	// Build the game.
	game := &domain.Game{
		BankerID:       data.bankerID,
		GameType:       data.gameType,
		Currency:       data.currency,
		MoneyModel:     data.moneyModel,
		ChipValue:      chipValue,
		BuyInAmount:    buyInAmount,
		RebuyAllowed:   data.rebuyAllowed,
		RebuyPrice:     rebuyPrice,
		MaxRebuys:      maxRebuys,
		Duration:       duration,
		StartTime:      startTime,
		MinPlayers:     minPlayers,
		MaxPlayers:     maxPlayers,
		RankingPrimary: data.rankingPrimary,
	}

	createdGame, err := b.svc.CreateGame(ctx, cb.From.ID, clubID, game)
	if err != nil {
		b.log.Warn("failed to create game", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка при создании игры: %v", err))
		return
	}

	b.setState(cb.From.ID, stateIdle, 0)

	// Send game menu to the private chat.
	b.sendTextWithKeyboard(cb.Message.Chat.ID, fmt.Sprintf("Игра [%s] создана (ID: %d). Статус: planned", gameTypeLabel(createdGame.GameType), createdGame.ID), b.gameMenuKeyboard(ctx, clubID, createdGame.ID, cb.From.ID))

	// Send notification to the club group chat.
	b.sendGroupNotification(ctx, clubID, fmt.Sprintf("🎲 Создана новая игра #%d %s", createdGame.ID, createdGame.StartTime.Format("02.01.2006 15:04")))

	// Send personal invitations to all active club members.
	b.sendGameInvitations(ctx, cb.From.ID, clubID, createdGame.ID)
}

// handleGameCreateCancel processes the "Отмена" button in the game creation screen.
func (b *Bot) handleGameCreateCancel(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, err := strconv.ParseInt(strings.TrimPrefix(cb.Data, cbGameCreateCancel+":"), 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}

	b.setState(cb.From.ID, stateIdle, 0)
	b.sendTextWithKeyboard(cb.Message.Chat.ID, "Создание игры отменено.", manageSubMenuKeyboard(clubID, ""))
}

// handleGameList displays the list of games for a club.
func (b *Bot) handleGameList(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, err := strconv.ParseInt(strings.TrimPrefix(cb.Data, cbGameList+":"), 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}

	games, err := b.svc.GetClubGames(ctx, cb.From.ID, clubID)
	if err != nil {
		b.log.Error("failed to get club games", "error", err)
		b.sendText(cb.Message.Chat.ID, "Ошибка при получении списка игр.")
		return
	}

	if len(games) == 0 {
		b.sendTextWithKeyboard(cb.Message.Chat.ID, "В клубе пока нет игр.", manageSubMenuKeyboard(clubID, ""))
		return
	}

	b.sendTextWithKeyboard(cb.Message.Chat.ID, "Игры клуба:", gameListKeyboard(clubID, games))
}

// handleSelectGame displays the game menu for a selected game.
func (b *Bot) handleSelectGame(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: select_game:<club_id>:<game_id>
	rest := strings.TrimPrefix(cb.Data, cbSelectGame+":")
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}

	b.setState(cb.From.ID, stateIdle, 0)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Игра:", b.gameMenuKeyboard(ctx, clubID, gameID, cb.From.ID))
}

// handleGameInfo displays detailed information about a game.
func (b *Bot) handleGameInfo(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_info:<club_id>:<game_id>
	rest := strings.TrimPrefix(cb.Data, cbGameInfo+":")
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}

	game, err := b.svc.GetGame(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.log.Error("failed to get game info", "error", err)
		b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Ошибка при получении информации об игре.", b.gameMenuKeyboard(ctx, clubID, gameID, cb.From.ID))
		return
	}

	text := formatGameInfo(game)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, b.gameMenuKeyboard(ctx, clubID, gameID, cb.From.ID))
}

// handleGameParticipants displays the list of game participants.
func (b *Bot) handleGameParticipants(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_participants:<club_id>:<game_id>
	rest := strings.TrimPrefix(cb.Data, cbGameParticipants+":")
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}

	participants, err := b.svc.GetGameParticipants(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.log.Error("failed to get game participants", "error", err)
		b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Ошибка при получении списка участников.", b.gameMenuKeyboard(ctx, clubID, gameID, cb.From.ID))
		return
	}

	if len(participants) == 0 {
		b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "В игре пока нет участников.", b.gameMenuKeyboard(ctx, clubID, gameID, cb.From.ID))
		return
	}

	var sb strings.Builder
	sb.WriteString("Участники игры:\n\n")
	for i, p := range participants {
		name := p.Player.FirstName
		if p.Player.LastName != "" {
			name += " " + p.Player.LastName
		}
		if p.Player.Nickname != "" && p.Player.Nickname != p.Player.FirstName {
			name += " (@" + p.Player.Nickname + ")"
		}
		sb.WriteString(fmt.Sprintf("%d. %s — %s\n", i+1, name, participantStatusLabel(p.Status)))
	}

	userRole := ""
	if role, err := b.svc.GetUserRole(ctx, cb.From.ID, clubID); err == nil {
		userRole = role
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, sb.String(), gameParticipantListKeyboard(clubID, gameID, participants, userRole))
}

// handleGameParticipantAction displays management options for a specific game participant.
func (b *Bot) handleGameParticipantAction(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: member_action:<club_id>:<game_id>:<player_id>
	rest := strings.TrimPrefix(cb.Data, cbMemberAction+":")
	parts := strings.SplitN(rest, ":", 3)
	if len(parts) != 3 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}
	playerID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игрока.")
		return
	}

	participants, err := b.svc.GetGameParticipants(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.log.Error("failed to get game participants", "error", err)
		b.sendText(cb.Message.Chat.ID, "Ошибка при получении данных участника.")
		return
	}

	var target *domain.GameParticipantWithPlayer
	for _, p := range participants {
		if p.PlayerID == playerID {
			target = p
			break
		}
	}
	if target == nil {
		b.sendText(cb.Message.Chat.ID, "Участник не найден.")
		return
	}

	userRole := ""
	if role, err := b.svc.GetUserRole(ctx, cb.From.ID, clubID); err == nil {
		userRole = role
	}

	name := target.Player.FirstName
	if target.Player.LastName != "" {
		name += " " + target.Player.LastName
	}
	if target.Player.Nickname != "" && target.Player.Nickname != target.Player.FirstName {
		name += " (@" + target.Player.Nickname + ")"
	}

	text := fmt.Sprintf("👤 %s\nСтатус: %s", name, participantStatusLabel(target.Status))
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, gameParticipantActionKeyboard(clubID, gameID, playerID, target, userRole))
}

// handleGameInvite processes the "Пригласить" button in the game menu.
func (b *Bot) handleGameInvite(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_invite:<club_id>:<game_id>
	rest := strings.TrimPrefix(cb.Data, cbGameInvite+":")
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}

	b.setState(cb.From.ID, stateGameInviteMember, clubID)
	b.states[cb.From.ID].gameID = gameID
	b.sendText(cb.Message.Chat.ID, "Введите Telegram username участника (например, @john_doe):")
}

// handleGameInviteMember processes the text input containing the username to invite to a game.
func (b *Bot) handleGameInviteMember(ctx context.Context, msg *tgbotapi.Message, clubID, gameID int64, username string) {
	username = strings.TrimSpace(username)
	username = strings.TrimPrefix(username, "@")
	if username == "" {
		b.sendText(msg.Chat.ID, "Username не может быть пустым. Попробуйте снова:")
		return
	}

	player, err := b.svc.GetPlayerByUsername(ctx, username)
	if err != nil {
		b.log.Warn("player not found by username for game invite", "username", username)
		b.sendTextWithKeyboard(msg.Chat.ID, fmt.Sprintf("Пользователь @%s не найден в системе.", username), b.gameMenuKeyboard(ctx, clubID, gameID, msg.From.ID))
		return
	}

	if player.TgUserID == nil {
		b.sendTextWithKeyboard(msg.Chat.ID, fmt.Sprintf("У пользователя @%s не привязан Telegram ID.", username), b.gameMenuKeyboard(ctx, clubID, gameID, msg.From.ID))
		return
	}

	// Verify the user is a member of this Telegram group.
	chatMember, err := b.api.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: msg.Chat.ID,
			UserID: *player.TgUserID,
		},
	})
	if err != nil {
		b.log.Error("failed to check chat member for game invite", "error", err, "tg_user_id", *player.TgUserID)
		b.sendTextWithKeyboard(msg.Chat.ID, "Ошибка при проверке участия пользователя в группе.", b.gameMenuKeyboard(ctx, clubID, gameID, msg.From.ID))
		return
	}

	if chatMember.HasLeft() || chatMember.WasKicked() {
		b.sendTextWithKeyboard(msg.Chat.ID, fmt.Sprintf("Пользователь @%s не является участником этой группы.", username), b.gameMenuKeyboard(ctx, clubID, gameID, msg.From.ID))
		return
	}

	_, game, err := b.svc.InviteToGame(ctx, msg.From.ID, clubID, gameID, *player.TgUserID)
	if err != nil {
		b.log.Warn("failed to invite member to game", "error", err)
		b.sendTextWithKeyboard(msg.Chat.ID, fmt.Sprintf("Ошибка: %v", err), b.gameMenuKeyboard(ctx, clubID, gameID, msg.From.ID))
		return
	}

	b.setState(msg.From.ID, stateIdle, 0)

	// Send personal invitation to the player.
	if player.TgUserID != nil {
		inviteText := fmt.Sprintf("Вас приглашают в игру [%s]. Примите участие или отклоните приглашение.", gameTypeLabel(game.GameType))
		b.sendTextWithKeyboard(*player.TgUserID, inviteText, gameAcceptDeclineKeyboard(clubID, game.ID))
	}

	b.sendTextWithKeyboard(msg.Chat.ID, fmt.Sprintf("Приглашение отправлено для @%s.", username), b.gameMenuKeyboard(ctx, clubID, gameID, msg.From.ID))
}

// handleGameAccept processes the player's acceptance of a game invitation.
func (b *Bot) handleGameAccept(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_accept:<club_id>:<game_id>
	rest := strings.TrimPrefix(cb.Data, cbGameAccept+":")
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}

	player, game, notifyIDs, err := b.svc.AcceptGameParticipation(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.log.Warn("failed to accept game participation", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	b.sendText(cb.Message.Chat.ID, "Вы приняли участие в игре. Ожидайте подтверждения от владельца или администратора.")

	// Notify owner and admin.
	for _, id := range notifyIDs {
		if id == cb.From.ID {
			continue
		}
		notifyText := fmt.Sprintf("Пользователь %s принял участие в игре [%s] (ID: %d). Подтвердить участие?", cb.From.FirstName, gameTypeLabel(game.GameType), game.ID)
		b.sendTextWithKeyboard(id, notifyText, gameConfirmParticipationKeyboard(clubID, gameID, player.ID))
	}
}

// handleGameDecline processes the player's decline of a game invitation.
func (b *Bot) handleGameDecline(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_decline:<club_id>:<game_id>
	rest := strings.TrimPrefix(cb.Data, cbGameDecline+":")
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}

	game, notifyIDs, err := b.svc.DeclineGameParticipation(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.log.Warn("failed to decline game participation", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	b.sendText(cb.Message.Chat.ID, "Вы отказались от участия в игре.")

	// Notify owner and admin.
	for _, id := range notifyIDs {
		if id == cb.From.ID {
			continue
		}
		notifyText := fmt.Sprintf("Пользователь %s отказался от участия в игре [%s] (ID: %d).", cb.From.FirstName, gameTypeLabel(game.GameType), game.ID)
		b.sendText(id, notifyText)
	}
}

// handleGameConfirmParticipation processes the owner/admin's confirmation of a player's participation.
func (b *Bot) handleGameConfirmParticipation(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_confirm_participation:<club_id>:<game_id>:<player_id>
	rest := strings.TrimPrefix(cb.Data, cbGameConfirmParticipation+":")
	parts := strings.SplitN(rest, ":", 3)
	if len(parts) != 3 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}
	playerID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игрока.")
		return
	}

	player, game, err := b.svc.ConfirmGameParticipation(ctx, cb.From.ID, clubID, gameID, playerID)
	if err != nil {
		b.log.Warn("failed to confirm game participation", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Notify the player.
	if player.TgUserID != nil {
		notifyText := fmt.Sprintf("Ваше участие в игре [%s] подтверждено. Добро пожаловать!", gameTypeLabel(game.GameType))
		b.sendText(*player.TgUserID, notifyText)
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Участие подтверждено.", b.gameMenuKeyboard(ctx, clubID, gameID, cb.From.ID))
}

// handleGameRemoveParticipant processes removing a participant from a game.
func (b *Bot) handleGameRemoveParticipant(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_remove_participant:<club_id>:<game_id>:<player_id>
	rest := strings.TrimPrefix(cb.Data, cbGameRemoveParticipant+":")
	parts := strings.SplitN(rest, ":", 3)
	if len(parts) != 3 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}
	playerID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игрока.")
		return
	}

	player, err := b.svc.RemoveGameParticipant(ctx, cb.From.ID, clubID, gameID, playerID)
	if err != nil {
		b.log.Warn("failed to remove game participant", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Notify the player.
	if player.TgUserID != nil {
		b.sendText(*player.TgUserID, "Вы удалены из игры.")
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Участник удален из игры.", b.gameMenuKeyboard(ctx, clubID, gameID, cb.From.ID))
}

// handleGameCancel processes the "Отменить игру" button.
func (b *Bot) handleGameCancel(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_cancel:<club_id>:<game_id>
	rest := strings.TrimPrefix(cb.Data, cbGameCancel+":")
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}

	b.sendTextWithKeyboard(cb.Message.Chat.ID, "Вы уверены, что хотите отменить игру?", gameConfirmCancelKeyboard(clubID, gameID))
}

// handleGameConfirmCancel processes the confirmation of game cancellation.
func (b *Bot) handleGameConfirmCancel(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_confirm_cancel:<club_id>:<game_id>
	rest := strings.TrimPrefix(cb.Data, cbGameConfirmCancel+":")
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}

	if err := b.svc.CancelGame(ctx, cb.From.ID, clubID, gameID); err != nil {
		b.log.Warn("failed to cancel game", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Notify all game participants.
	participants, err := b.svc.GetGameParticipants(ctx, cb.From.ID, clubID, gameID)
	if err == nil {
		for _, p := range participants {
			if p.Player.TgUserID != nil && *p.Player.TgUserID != cb.From.ID {
				b.sendText(*p.Player.TgUserID, "Игра отменена.")
			}
		}
	}

	// Send notification to the club group chat.
	game, _ := b.svc.GetGame(ctx, cb.From.ID, clubID, gameID)
	if game != nil {
		b.sendGroupNotification(ctx, clubID, fmt.Sprintf("🎲 Игра #%d %s отменена", gameID, game.StartTime.Format("02.01.2006 15:04")))
	}

	b.setState(cb.From.ID, stateIdle, 0)
	b.sendTextWithKeyboard(cb.Message.Chat.ID, "Игра отменена.", b.gameMenuKeyboard(ctx, clubID, gameID, cb.From.ID))
}

// handleGameChangeParams processes the "Изменить параметры" button.
func (b *Bot) handleGameChangeParams(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_change_params:<club_id>:<game_id>
	rest := strings.TrimPrefix(cb.Data, cbGameChangeParams+":")
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}

	game, err := b.svc.GetGame(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.log.Error("failed to get game for change params", "error", err)
		b.sendText(cb.Message.Chat.ID, "Ошибка при получении информации об игре.")
		return
	}

	// Initialize game creation data from existing game.
	data := &gameCreationData{
		gameType:       game.GameType,
		bankerID:       game.BankerID,
		bankerName:     "не выбран",
		currency:       game.Currency,
		moneyModel:     game.MoneyModel,
		chipValue:      strconv.FormatFloat(game.ChipValue, 'f', -1, 64),
		buyInAmount:    strconv.FormatFloat(game.BuyInAmount, 'f', -1, 64),
		rebuyAllowed:   game.RebuyAllowed,
		startTime:      game.StartTime.Format("02.01.2006 15:04"),
		minPlayers:     strconv.Itoa(game.MinPlayers),
		maxPlayers:     strconv.Itoa(game.MaxPlayers),
		rankingPrimary: game.RankingPrimary,
	}

	if game.RebuyPrice != nil {
		data.rebuyPrice = strconv.FormatFloat(*game.RebuyPrice, 'f', -1, 64)
	}
	if game.MaxRebuys != nil {
		data.maxRebuys = strconv.Itoa(*game.MaxRebuys)
	}
	if game.Duration != nil {
		data.duration = formatDuration(*game.Duration)
	}

	// Get banker name.
	members, err := b.svc.GetClubMembers(ctx, cb.From.ID, clubID)
	if err == nil {
		for _, m := range members {
			if m.ID == game.BankerID {
				data.bankerName = memberShortName(m.Player)
				break
			}
		}
	}

	b.setState(cb.From.ID, stateCreateGame, clubID)
	b.states[cb.From.ID].gameID = gameID
	b.states[cb.From.ID].gameData = data

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Изменение параметров игры:", gameCreationKeyboard(data, clubID, nil))
}

// handleGameBack processes the "Назад" button in game menus.
func (b *Bot) handleGameBack(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: game_back:<club_id>:<game_id>
	rest := strings.TrimPrefix(cb.Data, cbGameBack+":")
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}

	b.setState(cb.From.ID, stateIdle, 0)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Игра:", b.gameMenuKeyboard(ctx, clubID, gameID, cb.From.ID))
}

// sendGameInvitations sends personal messages to all active club members with
// Accept/Decline buttons for the game.
func (b *Bot) sendGameInvitations(ctx context.Context, tgUserID, clubID, gameID int64) {
	participants, err := b.svc.GetGameParticipants(ctx, tgUserID, clubID, gameID)
	if err != nil {
		b.log.Error("failed to get game participants for invitations", "error", err)
		return
	}

	for _, p := range participants {
		if p.Player.TgUserID != nil {
			inviteText := "Вас приглашают в игру. Примите участие или отклоните приглашение."
			b.sendTextWithKeyboard(*p.Player.TgUserID, inviteText, gameAcceptDeclineKeyboard(clubID, gameID))
		}
	}
}

// gameMenuKeyboard is a helper that wraps gameMenuKeyboard with role and banker lookup.
func (b *Bot) gameMenuKeyboard(ctx context.Context, clubID, gameID int64, tgUserID int64) tgbotapi.InlineKeyboardMarkup {
	userRole := ""
	if role, err := b.svc.GetUserRole(ctx, tgUserID, clubID); err == nil {
		userRole = role
	}

	game, err := b.svc.GetGame(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return gameMenuKeyboard(clubID, gameID, userRole, false, "planned")
	}

	isBanker := false
	player, pErr := b.svc.GetPlayerByTgUserID(ctx, tgUserID)
	if pErr == nil {
		member, mErr := b.svc.GetClubMember(ctx, clubID, player.ID)
		if mErr == nil && member.ID == game.BankerID {
			isBanker = true
		}
	}

	if game.Status == "active" {
		hasTimer := game.Duration != nil
		return activeGameMenuKeyboard(clubID, gameID, userRole, isBanker, game.GameType, hasTimer)
	}

	return gameMenuKeyboard(clubID, gameID, userRole, isBanker, game.Status)
}

// paramPrompt returns the prompt text for a given parameter.
func paramPrompt(param string) string {
	switch param {
	case "chip_value":
		return "Введите chip value (например, 1):"
	case "buy_in":
		return "Введите стоимость buy-in (например, 20):"
	case "rebuy_price":
		return "Введите стоимость rebuy (например, 20):"
	case "max_rebuys":
		return "Введите максимальное количество rebuy (например, 2):"
	case "start_time":
		return "Введите время начала (формат: 15.08.2026 19:00):"
	case "duration":
		return "Введите продолжительность (например, 5 часов). Отправьте '-' для пропуска:"
	case "min_players":
		return "Введите минимальное количество игроков (например, 4):"
	case "max_players":
		return "Введите максимальное количество игроков (например, 10):"
	case "ranking_primary":
		return "Введите способ определения победителя (например, profit):"
	default:
		return "Введите значение:"
	}
}

// parseDuration parses a duration string like "5 часов" or "2 hours" into a time.Duration.
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	parts := strings.Fields(s)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid duration format")
	}

	value, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid duration value")
	}

	unit := parts[1]
	switch unit {
	case "часов", "часа", "час", "hours", "hour", "h":
		return time.Duration(value) * time.Hour, nil
	case "минут", "минута", "минуты", "minutes", "minute", "min":
		return time.Duration(value) * time.Minute, nil
	case "дней", "день", "дня", "days", "day":
		return time.Duration(value) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown duration unit: %s", unit)
	}
}

// formatDuration formats a time.Duration into a human-readable string.
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%d часов %d минут", hours, minutes)
	}
	return fmt.Sprintf("%d минут", minutes)
}

// memberShortName returns a short name for a player.
func memberShortName(p domain.Player) string {
	name := p.FirstName
	if p.LastName != "" {
		name += " " + p.LastName
	}
	return name
}

// --- Phase 04: Active game handlers ---

// handleGameStart starts a planned game, registers buy-ins, and transitions to active.
func (b *Bot) handleGameStart(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	_, err = b.svc.StartGame(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка при запуске игры: %v", err))
		return
	}

	// Count confirmed/accepted participants.
	participants, _ := b.svc.GetGameParticipants(ctx, cb.From.ID, clubID, gameID)
	buyInRegistered := 0
	for _, p := range participants {
		if p.Status == "confirmed" || p.Status == "accepted" {
			buyInRegistered++
		}
	}

	text := fmt.Sprintf("🎲 Игра #%d начата!\n\nBuy-in зарегистрирован для %d/%d игроков.", gameID, buyInRegistered, len(participants))
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, gameCompleteBuyinKeyboard(clubID, gameID))

	b.log.Info("game started", "game_id", gameID, "club_id", clubID, "tg_user_id", cb.From.ID)
}

// handleGameCompleteBuyin transitions from the buy-in registration screen to the game monitor.
func (b *Bot) handleGameCompleteBuyin(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	game, participants, err := b.svc.GetGameMonitor(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	currentStacks, _ := b.svc.GetCurrentStacks(ctx, cb.From.ID, clubID, gameID)
	text := formatGameMonitorText(game, participants, currentStacks)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, gameMonitorKeyboard(clubID, gameID, true, game.GameType, game.Duration != nil))
}

// handleGameMonitor shows the game monitor for banker/owner/admin or player monitor for regular players.
func (b *Bot) handleGameMonitor(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	game, err := b.svc.GetGame(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Check if user is banker/owner/admin.
	isBankerOrAdmin := false
	player, pErr := b.svc.GetPlayerByTgUserID(ctx, cb.From.ID)
	if pErr == nil {
		member, mErr := b.svc.GetClubMember(ctx, clubID, player.ID)
		if mErr == nil && (member.Role == "owner" || member.Role == "admin" || member.ID == game.BankerID) {
			isBankerOrAdmin = true
		}
	}

	if isBankerOrAdmin {
		participants, err := b.svc.GetGameParticipants(ctx, cb.From.ID, clubID, gameID)
		if err != nil {
			b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
			return
		}
		currentStacks, _ := b.svc.GetCurrentStacks(ctx, cb.From.ID, clubID, gameID)
		text := formatGameMonitorText(game, participants, currentStacks)
		b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, gameMonitorKeyboard(clubID, gameID, true, game.GameType, game.Duration != nil))
	} else {
		// Player monitor.
		game, participant, err := b.svc.GetPlayerGameMonitor(ctx, cb.From.ID, clubID, gameID)
		if err != nil {
			b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
			return
		}

		currentStacks, _ := b.svc.GetCurrentStacks(ctx, cb.From.ID, clubID, gameID)
		var currentStack *float64
		if stack, ok := currentStacks[participant.PlayerID]; ok {
			currentStack = &stack
		}
		text := formatPlayerMonitorText(game, participant, currentStack)
		b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, playerMonitorKeyboard(clubID, gameID))
	}
}

// handleGameBank shows the current bank for the game.
func (b *Bot) handleGameBank(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	game, participants, err := b.svc.GetGameMonitor(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	text := formatGameBankText(game, participants)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, gameBankKeyboard(clubID, gameID))
}

// handleGameExpenses shows the player expenses for the game.
func (b *Bot) handleGameExpenses(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	game, participants, err := b.svc.GetGameMonitor(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	text := formatGameExpensesText(game, participants)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, gameExpensesKeyboard(clubID, gameID))
}

// handleGameRebuy shows the player selection for rebuy.
func (b *Bot) handleGameRebuy(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	participants, err := b.svc.GetGameParticipants(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	text := "Выберите игрока для Rebuy:"
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, rebuyPlayerSelectKeyboard(clubID, gameID, participants))
}

// handleGameRebuyPlayer shows the player info and confirm button for rebuy.
func (b *Bot) handleGameRebuyPlayer(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, playerID, err := parseCallbackData3(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	game, participants, err := b.svc.GetGameMonitor(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	var participant *domain.GameParticipantWithPlayer
	for _, p := range participants {
		if p.PlayerID == playerID {
			participant = p
			break
		}
	}
	if participant == nil {
		b.sendText(cb.Message.Chat.ID, "Игрок не найден в игре.")
		return
	}

	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}

	maxRebuysStr := "неограниченно"
	if game.MaxRebuys != nil {
		maxRebuysStr = strconv.Itoa(*game.MaxRebuys)
	}

	text := fmt.Sprintf("Игрок: %s\n\nRebuy: %s\nВыполнено: %d\nМаксимум: %s",
		memberShortName(participant.Player),
		strconv.FormatFloat(rebuyPrice, 'f', -1, 64),
		participant.RebuyCount,
		maxRebuysStr,
	)

	if game.MaxRebuys != nil && participant.RebuyCount >= *game.MaxRebuys {
		text += "\n\nМаксимальное количество Rebuy достигнуто."
		b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, rebuyPlayerSelectKeyboard(clubID, gameID, participants))
	} else {
		b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, rebuyConfirmKeyboard(clubID, gameID, playerID))
	}
}

// handleGameRebuyConfirm registers a rebuy for the selected player.
func (b *Bot) handleGameRebuyConfirm(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, playerID, err := parseCallbackData3(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	if err := b.svc.RegisterRebuy(ctx, cb.From.ID, clubID, gameID, playerID); err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка при регистрации Rebuy: %v", err))
		return
	}

	// Show updated player list.
	participants, _ := b.svc.GetGameParticipants(ctx, cb.From.ID, clubID, gameID)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "✅ Rebuy зарегистрирован.", rebuyPlayerSelectKeyboard(clubID, gameID, participants))
}

// handleGameRebuyManage shows the player selection for rebuy management.
func (b *Bot) handleGameRebuyManage(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	participants, err := b.svc.GetGameParticipants(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Filter to only players with rebuy_count > 0.
	var rebuyParticipants []*domain.GameParticipantWithPlayer
	for _, p := range participants {
		if p.RebuyCount > 0 {
			rebuyParticipants = append(rebuyParticipants, p)
		}
	}

	if len(rebuyParticipants) == 0 {
		b.sendText(cb.Message.Chat.ID, "Нет игроков с Rebuy.")
		return
	}

	text := "Выберите игрока для управления Rebuy:"
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, rebuyManagePlayerSelectKeyboard(clubID, gameID, rebuyParticipants))
}

// handleGameRebuyFixOp shows the list of rebuy events for a player.
func (b *Bot) handleGameRebuyFixOp(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, playerID, err := parseCallbackData3(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	events, err := b.svc.GetRebuyEvents(ctx, cb.From.ID, clubID, gameID, playerID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	if len(events) == 0 {
		b.sendText(cb.Message.Chat.ID, "Нет событий Rebuy для этого игрока.")
		return
	}

	text := "Выберите событие Rebuy для исправления:"
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, rebuyFixOpKeyboard(clubID, gameID, playerID, events))
}

// handleGameRebuyFixConfirm starts the rebuy fix flow by asking for new count.
func (b *Bot) handleGameRebuyFixConfirm(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, playerID, _, err := parseCallbackData4(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	b.setState(cb.From.ID, stateGameRebuyFixCount, clubID)
	b.mu.Lock()
	b.states[cb.From.ID].gameID = gameID
	b.states[cb.From.ID].playerID = playerID
	b.mu.Unlock()

	b.sendText(cb.Message.Chat.ID, "Введите новое количество Rebuy:")
}

// handleGameRebuyFixCountInput processes the rebuy count input and applies the fix.
func (b *Bot) handleGameRebuyFixCountInput(ctx context.Context, msg *tgbotapi.Message, state *userState) {
	count, err := strconv.Atoi(strings.TrimSpace(msg.Text))
	if err != nil || count < 0 {
		b.sendText(msg.Chat.ID, "Введите корректное число.")
		return
	}

	if err := b.svc.FixRebuy(ctx, msg.From.ID, state.clubID, state.gameID, state.playerID, count); err != nil {
		b.sendText(msg.Chat.ID, fmt.Sprintf("Ошибка при исправлении Rebuy: %v", err))
		return
	}

	b.setState(msg.From.ID, stateIdle, 0)
	b.sendText(msg.Chat.ID, "✅ Rebuy исправлен.")
}

// handleGameAddPlayer shows the club members available to add to the game.
func (b *Bot) handleGameAddPlayer(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	members, err := b.svc.GetClubMembers(ctx, cb.From.ID, clubID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Filter to active members not already in the game.
	participants, _ := b.svc.GetGameParticipants(ctx, cb.From.ID, clubID, gameID)
	participantMap := make(map[int64]bool)
	for _, p := range participants {
		participantMap[p.PlayerID] = true
	}

	var availableMembers []*domain.ClubMemberWithPlayer
	for _, m := range members {
		if m.Status == "active" && !participantMap[m.PlayerID] {
			availableMembers = append(availableMembers, m)
		}
	}

	if len(availableMembers) == 0 {
		b.sendText(cb.Message.Chat.ID, "Нет доступных игроков для добавления.")
		return
	}

	text := "Выберите игрока для добавления:"
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, addPlayerSelectKeyboard(clubID, gameID, availableMembers))
}

// handleGameAddPlayerConfirm adds the selected player to the game.
func (b *Bot) handleGameAddPlayerConfirm(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, playerID, err := parseCallbackData3(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	if err := b.svc.AddPlayerToGame(ctx, cb.From.ID, clubID, gameID, playerID); err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка при добавлении игрока: %v", err))
		return
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "✅ Игрок добавлен в игру.", addPlayerSelectKeyboard(clubID, gameID, nil))
}

// handleGamePauseTimer pauses the game timer.
func (b *Bot) handleGamePauseTimer(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	if err := b.svc.PauseTimer(ctx, cb.From.ID, clubID, gameID); err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка при паузе таймера: %v", err))
		return
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "⏸ Таймер приостановлен.", timerControlKeyboard(clubID, gameID, true))
}

// handleGameResumeTimer resumes the game timer.
func (b *Bot) handleGameResumeTimer(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	if err := b.svc.ResumeTimer(ctx, cb.From.ID, clubID, gameID); err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка при возобновлении таймера: %v", err))
		return
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "▶️ Таймер возобновлен.", timerControlKeyboard(clubID, gameID, false))
}

// handleGameExtend shows the extension duration options.
func (b *Bot) handleGameExtend(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Выберите продолжительность продления:", gameExtendSelectKeyboard(clubID, gameID))
}

// handleGameExtendSelect extends the game by the selected duration.
func (b *Bot) handleGameExtendSelect(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, minutes, err := parseCallbackData3Int(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	duration := time.Duration(minutes) * time.Minute
	if err := b.svc.ExtendGame(ctx, cb.From.ID, clubID, gameID, duration); err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка при продлении игры: %v", err))
		return
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, fmt.Sprintf("✅ Игра продлена на %d минут.", minutes), gameExtendKeyboard(clubID, gameID))
}

// handleGamePlayerStack asks the player for their current chip stack.
func (b *Bot) handleGamePlayerStack(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	b.setState(cb.From.ID, stateGamePlayerStack, clubID)
	b.mu.Lock()
	b.states[cb.From.ID].gameID = gameID
	b.mu.Unlock()

	b.sendText(cb.Message.Chat.ID, "Введите количество ваших фишек:")
}

// handleGamePlayerStats shows the player's statistics for the game.
func (b *Bot) handleGamePlayerStats(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	game, participant, err := b.svc.GetPlayerGameStats(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	currentStacks, _ := b.svc.GetCurrentStacks(ctx, cb.From.ID, clubID, gameID)
	var currentStack *float64
	if stack, ok := currentStacks[participant.PlayerID]; ok {
		currentStack = &stack
	}
	text := formatPlayerStatsText(game, participant, currentStack)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, playerStatsKeyboard(clubID, gameID))
}

// handleGameStats shows the game statistics for banker/owner/admin.
func (b *Bot) handleGameStats(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	game, participants, err := b.svc.GetGameStatistics(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	currentStacks, _ := b.svc.GetCurrentStacks(ctx, cb.From.ID, clubID, gameID)
	text := formatGameStatsText(game, participants, currentStacks)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, gameStatsKeyboard(clubID, gameID))
}

// handleGameActiveBack goes back to the game list.
func (b *Bot) handleGameActiveBack(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, _, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	games, err := b.svc.GetClubGames(ctx, cb.From.ID, clubID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	text := "Игры клуба:"
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, gameListKeyboard(clubID, games))
}

// handleGamePlayerStackInput processes the player's stack input.
func (b *Bot) handleGamePlayerStackInput(ctx context.Context, msg *tgbotapi.Message, state *userState) {
	stack, err := strconv.ParseFloat(strings.TrimSpace(msg.Text), 64)
	if err != nil || stack < 0 {
		b.sendText(msg.Chat.ID, "Введите корректное число.")
		return
	}

	if err := b.svc.UpdateCurrentStack(ctx, msg.From.ID, state.clubID, state.gameID, stack); err != nil {
		b.sendText(msg.Chat.ID, fmt.Sprintf("Ошибка при сохранении стека: %v", err))
		return
	}

	b.setState(msg.From.ID, stateIdle, 0)
	b.sendText(msg.Chat.ID, fmt.Sprintf("✅ Текущий стек сохранен: %s", strconv.FormatFloat(stack, 'f', -1, 64)))
}

// --- Phase 05: Game end handlers ---

// handleGameEnd shows the game end screen with the list of participants and
// their chips_end values, along with action buttons.
func (b *Bot) handleGameEnd(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	game, participants, err := b.svc.GetGameMonitor(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	allChipsEntered := true
	for _, p := range participants {
		if p.ChipsEnd == nil {
			allChipsEntered = false
			break
		}
	}

	text := formatGameEndText(game, participants)
	keyboard := gameEndKeyboard(clubID, gameID, allChipsEntered)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, keyboard)
}

// handleGameEndPlayer handles the player selection for entering chips_end.
// If a player_id is provided in the callback data, it asks for the chips_end
// value. Otherwise, it shows the player selection keyboard.
func (b *Bot) handleGameEndPlayer(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	parts := strings.Split(cb.Data, ":")
	// Format: game_end_player:club_id:game_id or game_end_player:club_id:game_id:player_id

	clubID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	gameID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игры.")
		return
	}

	if len(parts) >= 4 {
		// Player selected — ask for chips_end input.
		playerID, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игрока.")
			return
		}

		b.setState(cb.From.ID, stateGameEndChipsInput, clubID)
		b.mu.Lock()
		b.states[cb.From.ID].gameID = gameID
		b.states[cb.From.ID].playerID = playerID
		b.mu.Unlock()

		b.sendText(cb.Message.Chat.ID, "Введите итоговое количество фишек (chips_end):")
		return
	}

	// Show player selection.
	_, participants, err := b.svc.GetGameMonitor(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	text := "Выберите игрока для ввода chips_end:"
	keyboard := gameEndPlayerSelectKeyboard(clubID, gameID, participants)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, keyboard)
}

// handleGameEndChipsInput processes the chips_end value entered by the banker.
func (b *Bot) handleGameEndChipsInput(ctx context.Context, msg *tgbotapi.Message) {
	tgUserID := msg.From.ID

	b.mu.RLock()
	state, exists := b.states[tgUserID]
	b.mu.RUnlock()

	if !exists || state.action != stateGameEndChipsInput {
		b.sendText(msg.Chat.ID, "Введите /cancel для отмены.")
		return
	}

	chipsEnd, err := strconv.ParseFloat(msg.Text, 64)
	if err != nil {
		b.sendText(msg.Chat.ID, "Введите число (например: 2450):")
		return
	}

	if chipsEnd < 0 {
		b.sendText(msg.Chat.ID, "Количество фишек не может быть отрицательным. Введите число:")
		return
	}

	clubID := state.clubID
	gameID := state.gameID
	playerID := state.playerID

	if err := b.svc.SetChipsEnd(ctx, tgUserID, clubID, gameID, playerID, chipsEnd); err != nil {
		b.sendText(msg.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	b.setState(tgUserID, stateIdle, 0)

	// Refresh the end game screen.
	_, participants, err := b.svc.GetGameMonitor(ctx, tgUserID, clubID, gameID)
	if err != nil {
		b.sendText(msg.Chat.ID, fmt.Sprintf("✅ chips_end сохранен: %s", strconv.FormatFloat(chipsEnd, 'f', -1, 64)))
		return
	}

	allChipsEntered := true
	for _, p := range participants {
		if p.ChipsEnd == nil {
			allChipsEntered = false
			break
		}
	}

	keyboard := gameEndKeyboard(clubID, gameID, allChipsEntered)
	b.sendTextWithKeyboard(msg.Chat.ID, fmt.Sprintf("✅ chips_end сохранен: %s", strconv.FormatFloat(chipsEnd, 'f', -1, 64)), keyboard)
}

// handleGameEndCheckBank shows the bank check result for the game.
func (b *Bot) handleGameEndCheckBank(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	bankCheck, err := b.svc.CheckGameBank(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	text := formatGameBankCheckText(bankCheck)
	keyboard := gameEndKeyboard(clubID, gameID, false)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, keyboard)
}

// handleGameEndConfirm shows a confirmation prompt before finishing the game.
func (b *Bot) handleGameEndConfirm(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	// Check if all chips_end are entered.
	_, participants, err := b.svc.GetGameMonitor(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	allChipsEntered := true
	for _, p := range participants {
		if p.ChipsEnd == nil {
			allChipsEntered = false
			break
		}
	}

	if !allChipsEntered {
		b.sendText(cb.Message.Chat.ID, "Сначала введите chips_end для всех игроков.")
		return
	}

	text := "⚠️ Вы уверены, что хотите завершить игру?\n\nВсе результаты будут рассчитаны автоматически."
	keyboard := gameEndConfirmKeyboard(clubID, gameID)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, keyboard)
}

// handleGameEndFinish performs the actual game finish.
func (b *Bot) handleGameEndFinish(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	if err := b.svc.FinishGame(ctx, cb.From.ID, clubID, gameID); err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Get finished game results for notifications.
	game, participants, err := b.svc.GetFinishedGameResults(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.log.Warn("failed to get finished game results for notification", "error", err)
	}

	// Send group chat notification with results.
	if err == nil {
		groupText := formatFinishedGameGroupNotification(game, participants)
		b.sendGroupNotification(ctx, clubID, groupText)
	}

	// Send personal messages to players with their results.
	if err == nil {
		for _, p := range participants {
			if p.Player.TgUserID == nil {
				continue
			}
			personalText := formatFinishedGamePersonalResult(game, p)
			b.sendText(*p.Player.TgUserID, personalText)
		}
	}

	text := "✅ Игра завершена!\n\nРезультаты рассчитаны и сохранены."
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Результаты", fmt.Sprintf("%s:%s:%s", cbGameResults, strconv.FormatInt(clubID, 10), strconv.FormatInt(gameID, 10))),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("К игре", fmt.Sprintf("%s:%s:%s", cbGameActiveBack, strconv.FormatInt(clubID, 10), strconv.FormatInt(gameID, 10))),
		),
	))
}

// --- Phase 06: Statistics and game results handlers ---

// handlePlayerStats displays the requesting player's aggregate statistics.
func (b *Bot) handlePlayerStats(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, err := strconv.ParseInt(strings.TrimPrefix(cb.Data, cbPlayerStats+":"), 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}

	stats, err := b.svc.GetPlayerStatistics(ctx, cb.From.ID, clubID)
	if err != nil {
		b.log.Error("failed to get player statistics", "error", err)
		b.sendText(cb.Message.Chat.ID, "Ошибка при получении статистики.")
		return
	}

	text := formatPlayerStatistics(stats)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s", cbBackToClubMenu, strconv.FormatInt(clubID, 10))),
		),
	)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, keyboard)
}

// handleClubStats displays the club's aggregate statistics.
func (b *Bot) handleClubStats(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, err := strconv.ParseInt(strings.TrimPrefix(cb.Data, cbClubStats+":"), 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}

	stats, err := b.svc.GetClubStatistics(ctx, cb.From.ID, clubID)
	if err != nil {
		b.log.Error("failed to get club statistics", "error", err)
		b.sendText(cb.Message.Chat.ID, "Ошибка при получении статистики.")
		return
	}

	text := formatClubStatistics(stats)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s", cbBackToClubMenu, strconv.FormatInt(clubID, 10))),
		),
	)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, keyboard)
}

// handleGameResults displays the results of a finished game.
func (b *Bot) handleGameResults(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	clubID, gameID, err := parseCallbackData2(cb.Data)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные параметры.")
		return
	}

	game, participants, err := b.svc.GetFinishedGameResults(ctx, cb.From.ID, clubID, gameID)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	text := formatFinishedGameResults(game, participants)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("К игре", fmt.Sprintf("%s:%s:%s", cbGameActiveBack, strconv.FormatInt(clubID, 10), strconv.FormatInt(gameID, 10))),
		),
	)
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, keyboard)
}

// --- Phase 06: Text formatting helpers ---

// formatPlayerStatistics formats player aggregate statistics for display.
func formatPlayerStatistics(stats *domain.PlayerStatisticsView) string {
	var sb strings.Builder
	sb.WriteString("📊 Моя статистика\n\n")

	sb.WriteString(fmt.Sprintf("Всего игр: %d\n", stats.TotalGames))
	sb.WriteString(fmt.Sprintf("Побед: %d\n", stats.GamesWon))
	sb.WriteString(fmt.Sprintf("Winrate: %.1f%%\n", stats.Winrate))
	sb.WriteString(fmt.Sprintf("Среднее место: %.1f\n", stats.AvgPlace))
	sb.WriteString(fmt.Sprintf("Podiums: %d\n", stats.Podiums))
	sb.WriteString(fmt.Sprintf("ITM: %.1f%%\n", stats.ITM))
	sb.WriteString(fmt.Sprintf("ROI: %.1f%%\n", stats.ROI))

	sb.WriteString(fmt.Sprintf("\nВложения:\n"))
	sb.WriteString(fmt.Sprintf("  Buy-in: %s (%d)\n", formatFloat(stats.TotalBuyInAmount), stats.TotalBuyInCount))
	sb.WriteString(fmt.Sprintf("  Rebuy: %s (%d)\n", formatFloat(stats.TotalRebuyAmount), stats.TotalRebuysCount))
	sb.WriteString(fmt.Sprintf("  Всего вложено: %s\n", formatFloat(stats.TotalInvested)))

	sb.WriteString(fmt.Sprintf("\nФишки:\n"))
	sb.WriteString(fmt.Sprintf("  Всего фишек: %s\n", formatFloat(stats.TotalChips)))

	sb.WriteString(fmt.Sprintf("\nПрибыль:\n"))
	sb.WriteString(fmt.Sprintf("  Всего: %s\n", formatFloat(stats.TotalProfit)))
	sb.WriteString(fmt.Sprintf("  Лучшая игра: %s\n", formatFloat(stats.BiggestWin)))
	sb.WriteString(fmt.Sprintf("  Худшая игра: %s\n", formatFloat(stats.BiggestLoss)))

	return sb.String()
}

// formatClubStatistics formats club aggregate statistics for display.
func formatClubStatistics(stats *domain.ClubStatistics) string {
	var sb strings.Builder
	sb.WriteString("🏆 Статистика клуба\n\n")

	sb.WriteString(fmt.Sprintf("Участники: %d\n", stats.TotalMembers))
	sb.WriteString(fmt.Sprintf("Всего игр: %d\n", stats.TotalGames))
	sb.WriteString(fmt.Sprintf("  Cash: %d\n", stats.CashGames))
	sb.WriteString(fmt.Sprintf("  Tournament: %d\n", stats.TournamentGames))

	sb.WriteString(fmt.Sprintf("\nФинансы:\n"))
	sb.WriteString(fmt.Sprintf("  Buy-in: %s\n", formatFloat(stats.TotalBuyInAmount)))
	sb.WriteString(fmt.Sprintf("  Rebuy: %s\n", formatFloat(stats.TotalRebuyAmount)))
	sb.WriteString(fmt.Sprintf("  Банк: %s\n", formatFloat(stats.TotalBank)))

	sb.WriteString(fmt.Sprintf("\nСредняя длительность игры: %s\n", formatDuration(stats.AverageGameDuration)))

	return sb.String()
}

// formatFinishedGameResults formats the results of a finished game for display.
func formatFinishedGameResults(game *domain.Game, participants []*domain.GameParticipantWithPlayer) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🎲 Игра #%d — результаты\n", game.ID))
	sb.WriteString(fmt.Sprintf("Тип: %s\n", gameTypeLabel(game.GameType)))
	sb.WriteString(fmt.Sprintf("Статус: %s\n", game.Status))
	sb.WriteString(fmt.Sprintf("Начало: %s\n", game.StartTime.Format("02.01.2006 15:04")))
	if game.EndTime != nil {
		sb.WriteString(fmt.Sprintf("Конец: %s\n", game.EndTime.Format("02.01.2006 15:04")))
	}

	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}

	totalBank := 0.0
	totalPayout := 0.0

	sb.WriteString(fmt.Sprintf("\nУчастники (%d):\n", len(participants)))
	for _, p := range participants {
		name := memberShortName(p.Player)
		buyInAmount := float64(p.BuyInCount) * game.BuyInAmount
		rebuyAmount := float64(p.RebuyCount) * rebuyPrice
		totalInvested := buyInAmount + rebuyAmount

		var payoutAmount float64
		if p.PayoutAmount != nil {
			payoutAmount = *p.PayoutAmount
		}
		profit := payoutAmount - totalInvested

		var roi float64
		if totalInvested > 0 {
			roi = (profit / totalInvested) * 100
		}

		totalBank += totalInvested
		totalPayout += payoutAmount

		placeStr := "—"
		if p.Place != nil {
			placeStr = strconv.Itoa(*p.Place)
		}

		sb.WriteString(fmt.Sprintf("  %s — место: %s, вложено: %s, выплата: %s, прибыль: %s, ROI: %.1f%%\n",
			name, placeStr, formatFloat(totalInvested), formatFloat(payoutAmount), formatFloat(profit), roi))
	}

	sb.WriteString(fmt.Sprintf("\nБанк: %s\n", formatFloat(totalBank)))
	sb.WriteString(fmt.Sprintf("Выплаты: %s\n", formatFloat(totalPayout)))

	return sb.String()
}

// formatFinishedGameGroupNotification formats the group chat notification text
// sent after a game finishes.
func formatFinishedGameGroupNotification(game *domain.Game, participants []*domain.GameParticipantWithPlayer) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("✅ Игра #%d завершена!\n\n", game.ID))
	sb.WriteString(fmt.Sprintf("Тип: %s\n", gameTypeLabel(game.GameType)))
	sb.WriteString(fmt.Sprintf("Начало: %s\n", game.StartTime.Format("02.01.2006 15:04")))
	if game.EndTime != nil {
		sb.WriteString(fmt.Sprintf("Конец: %s\n", game.EndTime.Format("02.01.2006 15:04")))
	}

	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}

	totalBank := 0.0
	totalPayout := 0.0

	sb.WriteString(fmt.Sprintf("\nРезультаты (%d участников):\n", len(participants)))
	for _, p := range participants {
		name := memberShortName(p.Player)
		buyInAmount := float64(p.BuyInCount) * game.BuyInAmount
		rebuyAmount := float64(p.RebuyCount) * rebuyPrice
		totalInvested := buyInAmount + rebuyAmount

		var payoutAmount float64
		if p.PayoutAmount != nil {
			payoutAmount = *p.PayoutAmount
		}
		profit := payoutAmount - totalInvested

		totalBank += totalInvested
		totalPayout += payoutAmount

		placeStr := "—"
		if p.Place != nil {
			placeStr = strconv.Itoa(*p.Place)
		}

		sb.WriteString(fmt.Sprintf("  %s — место: %s, прибыль: %s\n",
			name, placeStr, formatFloat(profit)))
	}

	sb.WriteString(fmt.Sprintf("\nБанк: %s\n", formatFloat(totalBank)))
	sb.WriteString(fmt.Sprintf("Выплаты: %s\n", formatFloat(totalPayout)))

	return sb.String()
}

// formatFinishedGamePersonalResult formats the personal result message sent
// to a player after a game finishes.
func formatFinishedGamePersonalResult(game *domain.Game, p *domain.GameParticipantWithPlayer) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🎲 Игра #%d завершена!\n\n", game.ID))
	sb.WriteString(fmt.Sprintf("Тип: %s\n", gameTypeLabel(game.GameType)))

	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}

	buyInAmount := float64(p.BuyInCount) * game.BuyInAmount
	rebuyAmount := float64(p.RebuyCount) * rebuyPrice
	totalInvested := buyInAmount + rebuyAmount

	var payoutAmount float64
	if p.PayoutAmount != nil {
		payoutAmount = *p.PayoutAmount
	}
	profit := payoutAmount - totalInvested

	var roi float64
	if totalInvested > 0 {
		roi = (profit / totalInvested) * 100
	}

	placeStr := "—"
	if p.Place != nil {
		placeStr = strconv.Itoa(*p.Place)
	}

	sb.WriteString(fmt.Sprintf("Мои результаты:\n"))
	sb.WriteString(fmt.Sprintf("  Место: %s\n", placeStr))
	sb.WriteString(fmt.Sprintf("  Buy-in: %s (%d)\n", formatFloat(buyInAmount), p.BuyInCount))
	sb.WriteString(fmt.Sprintf("  Rebuy: %s (%d)\n", formatFloat(rebuyAmount), p.RebuyCount))
	sb.WriteString(fmt.Sprintf("  Вложено: %s\n", formatFloat(totalInvested)))
	sb.WriteString(fmt.Sprintf("  Выплата: %s\n", formatFloat(payoutAmount)))
	sb.WriteString(fmt.Sprintf("  Прибыль: %s\n", formatFloat(profit)))
	sb.WriteString(fmt.Sprintf("  ROI: %.1f%%\n", roi))

	return sb.String()
}

// formatGameMonitorText formats the game monitor text for banker/owner/admin.
func formatGameMonitorText(game *domain.Game, participants []*domain.GameParticipantWithPlayer, currentStacks map[int64]float64) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🎲 Игра #%d\n", game.ID))
	sb.WriteString(fmt.Sprintf("Тип: %s\n", gameTypeLabel(game.GameType)))
	sb.WriteString(fmt.Sprintf("Статус: %s\n", game.Status))
	sb.WriteString(fmt.Sprintf("Начало: %s\n", game.StartTime.Format("02.01.2006 15:04")))

	if game.Duration != nil {
		timerStr := formatTimerDisplay(game)
		sb.WriteString(fmt.Sprintf("⏱ Время: %s\n", timerStr))
	}

	buyInRegistered := 0
	totalBuyIn := 0.0
	totalRebuy := 0.0
	totalBank := 0.0

	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}

	for _, p := range participants {
		if p.BuyInCount > 0 {
			buyInRegistered++
		}
		playerBuyIn := float64(p.BuyInCount) * game.BuyInAmount
		playerRebuy := float64(p.RebuyCount) * rebuyPrice
		totalBuyIn += playerBuyIn
		totalRebuy += playerRebuy
		totalBank += playerBuyIn + playerRebuy
	}

	sb.WriteString(fmt.Sprintf("\nУчастников: %d\n", len(participants)))
	sb.WriteString(fmt.Sprintf("Buy-in зарегистрирован: %d/%d\n", buyInRegistered, len(participants)))
	sb.WriteString(fmt.Sprintf("\nБанк: %s\n", formatFloat(totalBank)))
	sb.WriteString(fmt.Sprintf("Buy-in: %s\n", formatFloat(totalBuyIn)))
	sb.WriteString(fmt.Sprintf("Rebuy: %s\n", formatFloat(totalRebuy)))

	sb.WriteString("\nИгроки:\n")
	for _, p := range participants {
		name := memberShortName(p.Player)
		playerBuyIn := float64(p.BuyInCount) * game.BuyInAmount
		playerRebuy := float64(p.RebuyCount) * rebuyPrice
		stackStr := "—"
		if stack, ok := currentStacks[p.PlayerID]; ok {
			stackStr = formatFloat(stack)
		}
		sb.WriteString(fmt.Sprintf("%s  %s  %s\n", name, formatFloat(playerBuyIn+playerRebuy), stackStr))
	}

	return sb.String()
}

// formatPlayerMonitorText formats the player's game monitor text.
func formatPlayerMonitorText(game *domain.Game, participant *domain.GameParticipantWithPlayer, currentStack *float64) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🎲 Игра #%d\n", game.ID))
	sb.WriteString(fmt.Sprintf("Тип: %s\n", gameTypeLabel(game.GameType)))
	sb.WriteString(fmt.Sprintf("Статус: %s\n", game.Status))

	if game.Duration != nil {
		timerStr := formatTimerDisplay(game)
		sb.WriteString(fmt.Sprintf("⏱ Время: %s\n", timerStr))
	}

	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}

	playerBuyIn := float64(participant.BuyInCount) * game.BuyInAmount
	playerRebuy := float64(participant.RebuyCount) * rebuyPrice
	totalInvested := playerBuyIn + playerRebuy

	sb.WriteString(fmt.Sprintf("\nМои вложения:\n"))
	sb.WriteString(fmt.Sprintf("Buy-in: %s\n", formatFloat(playerBuyIn)))
	sb.WriteString(fmt.Sprintf("Rebuy: %s\n", formatFloat(playerRebuy)))
	sb.WriteString(fmt.Sprintf("Всего: %s\n", formatFloat(totalInvested)))
	sb.WriteString(fmt.Sprintf("\nRebuy: %d\n", participant.RebuyCount))

	stackStr := "—"
	if currentStack != nil {
		stackStr = formatFloat(*currentStack)
	}
	sb.WriteString(fmt.Sprintf("Текущий стек: %s\n", stackStr))

	return sb.String()
}

// formatGameBankText formats the current bank text.
func formatGameBankText(game *domain.Game, participants []*domain.GameParticipantWithPlayer) string {
	totalBuyIn := 0.0
	totalRebuy := 0.0
	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}
	for _, p := range participants {
		totalBuyIn += float64(p.BuyInCount) * game.BuyInAmount
		totalRebuy += float64(p.RebuyCount) * rebuyPrice
	}
	totalBank := totalBuyIn + totalRebuy

	var sb strings.Builder
	sb.WriteString("💰 Банк игры\n\n")
	sb.WriteString(fmt.Sprintf("Всего: %s\n", formatFloat(totalBank)))
	sb.WriteString(fmt.Sprintf("Buy-in: %s\n", formatFloat(totalBuyIn)))
	sb.WriteString(fmt.Sprintf("Rebuy: %s\n", formatFloat(totalRebuy)))
	return sb.String()
}

// formatGameExpensesText formats the player expenses text.
func formatGameExpensesText(game *domain.Game, participants []*domain.GameParticipantWithPlayer) string {
	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}

	var sb strings.Builder
	sb.WriteString("💸 Расходы игроков\n\n")
	for _, p := range participants {
		name := memberShortName(p.Player)
		playerBuyIn := float64(p.BuyInCount) * game.BuyInAmount
		playerRebuy := float64(p.RebuyCount) * rebuyPrice
		total := playerBuyIn + playerRebuy
		sb.WriteString(fmt.Sprintf("%s\n", name))
		sb.WriteString(fmt.Sprintf("  Buy-in: %s\n", formatFloat(playerBuyIn)))
		sb.WriteString(fmt.Sprintf("  Rebuy: %s\n", formatFloat(playerRebuy)))
		sb.WriteString(fmt.Sprintf("  Rebuy count: %d\n", p.RebuyCount))
		sb.WriteString(fmt.Sprintf("  Всего: %s\n\n", formatFloat(total)))
	}
	return sb.String()
}

// formatPlayerStatsText formats the player's statistics text.
func formatPlayerStatsText(game *domain.Game, participant *domain.GameParticipant, currentStack *float64) string {
	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}
	playerBuyIn := float64(participant.BuyInCount) * game.BuyInAmount
	playerRebuy := float64(participant.RebuyCount) * rebuyPrice
	totalInvested := playerBuyIn + playerRebuy

	var sb strings.Builder
	sb.WriteString("📊 Моя статистика\n\n")
	sb.WriteString(fmt.Sprintf("Buy-in count: %d\n", participant.BuyInCount))
	sb.WriteString(fmt.Sprintf("Buy-in amount: %s\n", formatFloat(playerBuyIn)))
	sb.WriteString(fmt.Sprintf("Rebuy count: %d\n", participant.RebuyCount))
	sb.WriteString(fmt.Sprintf("Rebuy amount: %s\n", formatFloat(playerRebuy)))
	sb.WriteString(fmt.Sprintf("Total invested: %s\n", formatFloat(totalInvested)))

	stackStr := "—"
	if currentStack != nil {
		stackStr = formatFloat(*currentStack)
	}
	sb.WriteString(fmt.Sprintf("Current stack: %s\n", stackStr))
	return sb.String()
}

// formatGameStatsText formats the game statistics text.
func formatGameStatsText(game *domain.Game, participants []*domain.GameParticipantWithPlayer, currentStacks map[int64]float64) string {
	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🎲 Игра #%d\n\n", game.ID))

	buyInRegistered := 0
	for _, p := range participants {
		if p.BuyInCount > 0 {
			buyInRegistered++
		}
	}

	sb.WriteString(fmt.Sprintf("Участников: %d\n", len(participants)))
	sb.WriteString(fmt.Sprintf("Buy-in зарегистрирован: %d/%d\n", buyInRegistered, len(participants)))

	totalBuyIn := 0.0
	totalRebuy := 0.0
	for _, p := range participants {
		totalBuyIn += float64(p.BuyInCount) * game.BuyInAmount
		totalRebuy += float64(p.RebuyCount) * rebuyPrice
	}
	totalBank := totalBuyIn + totalRebuy

	sb.WriteString(fmt.Sprintf("Банк: %s\n", formatFloat(totalBank)))
	sb.WriteString(fmt.Sprintf("Buy-in: %s\n", formatFloat(totalBuyIn)))
	sb.WriteString(fmt.Sprintf("Rebuy: %s\n", formatFloat(totalRebuy)))

	sb.WriteString("\nИгроки:\n")
	for _, p := range participants {
		name := memberShortName(p.Player)
		playerBuyIn := float64(p.BuyInCount) * game.BuyInAmount
		playerRebuy := float64(p.RebuyCount) * rebuyPrice
		stackStr := "—"
		if stack, ok := currentStacks[p.PlayerID]; ok {
			stackStr = formatFloat(stack)
		}
		sb.WriteString(fmt.Sprintf("%s  %s  %s\n", name, formatFloat(playerBuyIn+playerRebuy), stackStr))
	}

	if game.Duration != nil {
		sb.WriteString(fmt.Sprintf("\nТаймер: %s\n", formatTimerDisplay(game)))
	}

	return sb.String()
}

// formatFloat formats a float64 for display, trimming trailing zeros.
func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// playerMonitorKeyboard returns the keyboard for the player's game monitor view.
func playerMonitorKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Моя статистика", fmt.Sprintf("%s:%s:%s", cbGamePlayerStats, cid, gid)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Ввести текущий стек", fmt.Sprintf("%s:%s:%s", cbGamePlayerStack, cid, gid)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameActiveBack, cid, gid)),
		),
	)
}

// parseCallbackData2 parses a callback data string with 2 integer parts after the prefix.
// Format: prefix:<int>:<int>
func parseCallbackData2(data string) (int64, int64, error) {
	parts := strings.Split(data, ":")
	if len(parts) < 3 {
		return 0, 0, fmt.Errorf("invalid callback data")
	}
	clubID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	gameID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return clubID, gameID, nil
}

// parseCallbackData3 parses a callback data string with 3 integer parts after the prefix.
// Format: prefix:<int>:<int>:<int>
func parseCallbackData3(data string) (int64, int64, int64, error) {
	parts := strings.Split(data, ":")
	if len(parts) < 4 {
		return 0, 0, 0, fmt.Errorf("invalid callback data")
	}
	clubID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	gameID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	playerID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	return clubID, gameID, playerID, nil
}

// parseCallbackData3Int parses a callback data string with 2 integer parts and 1 int part after the prefix.
// Format: prefix:<int>:<int>:<int>
func parseCallbackData3Int(data string) (int64, int64, int, error) {
	parts := strings.Split(data, ":")
	if len(parts) < 4 {
		return 0, 0, 0, fmt.Errorf("invalid callback data")
	}
	clubID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	gameID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	minutes, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, 0, 0, err
	}
	return clubID, gameID, minutes, nil
}

// parseCallbackData4 parses a callback data string with 4 integer parts after the prefix.
// Format: prefix:<int>:<int>:<int>:<int>
func parseCallbackData4(data string) (int64, int64, int64, int64, error) {
	parts := strings.Split(data, ":")
	if len(parts) < 5 {
		return 0, 0, 0, 0, fmt.Errorf("invalid callback data")
	}
	clubID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	gameID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	playerID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	eventID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return clubID, gameID, playerID, eventID, nil
}

// formatGameInfo returns a formatted string with game information.
func formatGameInfo(game *domain.Game) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🏠 Игра #%d\n", game.ID))
	sb.WriteString(fmt.Sprintf("Тип: %s\n", gameTypeLabel(game.GameType)))
	sb.WriteString(fmt.Sprintf("Статус: %s\n", game.Status))
	sb.WriteString(fmt.Sprintf("Валюта: %s\n", game.Currency))
	sb.WriteString(fmt.Sprintf("Денежная модель: %s\n", game.MoneyModel))
	sb.WriteString(fmt.Sprintf("Chip value: %s\n", strconv.FormatFloat(game.ChipValue, 'f', -1, 64)))
	sb.WriteString(fmt.Sprintf("Buy-in: %s\n", strconv.FormatFloat(game.BuyInAmount, 'f', -1, 64)))
	sb.WriteString(fmt.Sprintf("Rebuy: %s\n", boolYesNo(game.RebuyAllowed)))
	if game.RebuyPrice != nil {
		sb.WriteString(fmt.Sprintf("Rebuy price: %s\n", strconv.FormatFloat(*game.RebuyPrice, 'f', -1, 64)))
	}
	if game.MaxRebuys != nil {
		sb.WriteString(fmt.Sprintf("Max rebuys: %d\n", *game.MaxRebuys))
	}
	sb.WriteString(fmt.Sprintf("Start: %s\n", game.StartTime.Format("02.01.2006 15:04")))
	if game.Duration != nil {
		sb.WriteString(fmt.Sprintf("Duration: %s\n", formatDuration(*game.Duration)))
	}
	sb.WriteString(fmt.Sprintf("Min players: %d\n", game.MinPlayers))
	sb.WriteString(fmt.Sprintf("Max players: %d\n", game.MaxPlayers))
	sb.WriteString(fmt.Sprintf("Ranking: %s\n", game.RankingPrimary))
	return sb.String()
}

// --- Phase 05: Text formatting helpers ---

// formatGameEndText formats the game end screen text with participant list
// and their chips_end values.
func formatGameEndText(game *domain.Game, participants []*domain.GameParticipantWithPlayer) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🎲 Игра #%d\n", game.ID))
	sb.WriteString(fmt.Sprintf("Тип: %s\n", gameTypeLabel(game.GameType)))
	sb.WriteString(fmt.Sprintf("Статус: %s\n", game.Status))
	sb.WriteString(fmt.Sprintf("\nУчастники (%d):\n", len(participants)))
	for _, p := range participants {
		name := memberShortName(p.Player)
		chipsStr := "—"
		if p.ChipsEnd != nil {
			chipsStr = strconv.FormatFloat(*p.ChipsEnd, 'f', -1, 64)
		}
		sb.WriteString(fmt.Sprintf("  %s — chips_end: %s\n", name, chipsStr))
	}
	return sb.String()
}

// formatGameBankCheckText formats the bank check result text.
func formatGameBankCheckText(bankCheck *service.GameBankCheck) string {
	var sb strings.Builder
	sb.WriteString("💰 Проверка банка\n\n")
	sb.WriteString(fmt.Sprintf("Банк: %s\n", strconv.FormatFloat(bankCheck.TotalBank, 'f', -1, 64)))
	sb.WriteString(fmt.Sprintf("Выплата: %s\n", strconv.FormatFloat(bankCheck.TotalPayout, 'f', -1, 64)))
	if bankCheck.Mismatch {
		sb.WriteString(fmt.Sprintf("\n⚠️ Суммы не сходятся! Разница: %s\n", strconv.FormatFloat(bankCheck.Difference, 'f', -1, 64)))
		sb.WriteString("Разница будет распределена между игроками с положительным profit при завершении.")
	} else {
		sb.WriteString("\n✅ Суммы совпадают.")
	}
	return sb.String()
}
