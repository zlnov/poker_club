package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"poker-club/backend/internal/domain"
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

	// Send notification to the club group.
	b.sendTextWithKeyboard(cb.Message.Chat.ID, fmt.Sprintf("Игра [%s] создана (ID: %d). Статус: planned", gameTypeLabel(createdGame.GameType), createdGame.ID), b.gameMenuKeyboard(ctx, clubID, createdGame.ID, cb.From.ID))

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

// gameMenuKeyboard is a helper that wraps gameMenuKeyboard with role lookup.
func (b *Bot) gameMenuKeyboard(ctx context.Context, clubID, gameID int64, tgUserID int64) tgbotapi.InlineKeyboardMarkup {
	userRole := ""
	if role, err := b.svc.GetUserRole(ctx, tgUserID, clubID); err == nil {
		userRole = role
	}
	return gameMenuKeyboard(clubID, gameID, userRole)
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
