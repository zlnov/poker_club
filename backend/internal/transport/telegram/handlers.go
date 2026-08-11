package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"poker-club/backend/internal/domain"
)

// handleCommand processes text commands like /start.
func (b *Bot) handleCommand(ctx context.Context, update tgbotapi.Update) {
	msg := update.Message
	if msg == nil || msg.Text == "" {
		return
	}

	// Extract command (first word, without bot suffix).
	parts := strings.Fields(msg.Text)
	if len(parts) == 0 {
		return
	}
	cmd := strings.SplitN(parts[0], "@", 2)[0]

	switch cmd {
	case "/start":
		b.handleStart(ctx, msg)
	case "/bind":
		b.handleBind(ctx, msg)
	case "/invite":
		b.handleInviteCommand(ctx, msg)
	default:
		b.sendText(msg.Chat.ID, "Неизвестная команда. Используйте кнопки меню.")
	}
}

// handleStart processes the /start command — shows the main menu.
// If a deep link parameter is present (e.g. "invite_<club_id>"), it shows the
// invitation message instead of the main menu. The deep link does NOT create
// a new Player; it only works with an existing player and club_member.
func (b *Bot) handleStart(ctx context.Context, msg *tgbotapi.Message) {
	args := msg.CommandArguments()

	// Deep link: invite_<club_id>
	if strings.HasPrefix(args, "invite_") {
		b.handleStartWithInvite(ctx, msg, args)
		return
	}

	firstName := msg.From.FirstName
	lastName := msg.From.LastName
	nickname := msg.From.UserName
	if nickname == "" {
		nickname = firstName
	}

	_, err := b.svc.RegisterTelegramUser(ctx, msg.From.ID, firstName, lastName, nickname)
	if err != nil {
		b.log.Error("failed to register player on /start", "error", err)
		b.sendText(msg.Chat.ID, "Ошибка при регистрации. Попробуйте позже.")
		return
	}

	text := fmt.Sprintf("Добро пожаловать, %s!\n\nВыберите действие:", firstName)
	b.sendTextWithKeyboard(msg.Chat.ID, text, mainMenuKeyboardMarkup())
}

// handleStartWithInvite processes a /start command with a deep link parameter
// (e.g. "invite_<club_id>"). It shows the invitation message to the user
// without creating a new Player.
func (b *Bot) handleStartWithInvite(ctx context.Context, msg *tgbotapi.Message, args string) {
	clubID, err := strconv.ParseInt(strings.TrimPrefix(args, "invite_"), 10, 64)
	if err != nil {
		b.sendText(msg.Chat.ID, "Ошибка: неверная ссылка приглашения.")
		return
	}

	// Find the player by tg_user_id — do NOT create a new player.
	player, err := b.svc.GetPlayerByTgUserID(ctx, msg.From.ID)
	if err != nil {
		b.sendText(msg.Chat.ID, "Ошибка: ваш профиль не найден. Добавьтесь в группу клуба, чтобы получить приглашение.")
		return
	}

	// Find the existing club_member (pending invitation).
	_, err = b.svc.GetClubMember(ctx, clubID, player.ID)
	if err != nil {
		b.sendText(msg.Chat.ID, "Приглашение не найдено.")
		return
	}

	club, err := b.svc.GetClubInfo(ctx, clubID)
	if err != nil {
		b.sendText(msg.Chat.ID, "Ошибка при получении информации о клубе.")
		return
	}

	inviteText := fmt.Sprintf("Вас приглашают вступить в клуб «%s»", club.Name)
	b.sendTextWithKeyboard(msg.Chat.ID, inviteText, invitationKeyboard(club.ID))
}

// handleCallback processes inline keyboard button presses.
func (b *Bot) handleCallback(ctx context.Context, update tgbotapi.Update) {
	cb := update.CallbackQuery
	if cb == nil {
		return
	}

	// Always answer the callback to remove the "loading" state.
	_, _ = b.api.Request(tgbotapi.NewCallback(cb.ID, ""))

	data := cb.Data
	chatID := cb.Message.Chat.ID
	msgID := cb.Message.MessageID

	switch {
	case data == cbCreateClub:
		b.setState(cb.From.ID, stateCreateClub, 0)
		b.sendText(chatID, "Введите название клуба:")

	case data == cbMyClubs:
		b.showMyClubs(ctx, chatID, cb.From.ID)

	case strings.HasPrefix(data, cbSelectClub+":"):
		clubID, err := strconv.ParseInt(strings.TrimPrefix(data, cbSelectClub+":"), 10, 64)
		if err != nil {
			b.sendText(chatID, "Ошибка: неверный идентификатор клуба.")
			return
		}
		b.showClubMenu(chatID, msgID, clubID)

	case strings.HasPrefix(data, cbClubInfo+":"):
		clubID, err := strconv.ParseInt(strings.TrimPrefix(data, cbClubInfo+":"), 10, 64)
		if err != nil {
			b.sendText(chatID, "Ошибка: неверный идентификатор клуба.")
			return
		}
		b.showClubInfo(ctx, chatID, msgID, clubID)

	case strings.HasPrefix(data, cbChangeName+":"):
		clubID, err := strconv.ParseInt(strings.TrimPrefix(data, cbChangeName+":"), 10, 64)
		if err != nil {
			b.sendText(chatID, "Ошибка: неверный идентификатор клуба.")
			return
		}
		b.setState(cb.From.ID, stateChangeName, clubID)
		b.sendText(chatID, "Введите новое название клуба:")

	case strings.HasPrefix(data, cbCloseClub+":"):
		clubID, err := strconv.ParseInt(strings.TrimPrefix(data, cbCloseClub+":"), 10, 64)
		if err != nil {
			b.sendText(chatID, "Ошибка: неверный идентификатор клуба.")
			return
		}
		b.setState(cb.From.ID, stateCloseConfirm, clubID)
		b.sendTextWithKeyboard(chatID, "Вы уверены, что хотите закрыть клуб?", confirmCloseKeyboard(clubID))

	case strings.HasPrefix(data, cbConfirmClose+":"):
		clubID, err := strconv.ParseInt(strings.TrimPrefix(data, cbConfirmClose+":"), 10, 64)
		if err != nil {
			b.sendText(chatID, "Ошибка: неверный идентификатор клуба.")
			return
		}
		b.setState(cb.From.ID, stateIdle, 0)
		if err := b.svc.CloseClub(ctx, cb.From.ID, clubID); err != nil {
			b.sendText(chatID, fmt.Sprintf("Ошибка: %v", err))
			return
		}
		b.sendTextWithKeyboard(chatID, "Клуб закрыт.", mainMenuKeyboardMarkup())

	case strings.HasPrefix(data, cbCancelClose+":"):
		clubID, err := strconv.ParseInt(strings.TrimPrefix(data, cbCancelClose+":"), 10, 64)
		if err != nil {
			b.sendText(chatID, "Ошибка: неверный идентификатор клуба.")
			return
		}
		b.setState(cb.From.ID, stateIdle, 0)
		b.showClubMenu(chatID, msgID, clubID)

	case data == cbBackMain:
		b.setState(cb.From.ID, stateIdle, 0)
		b.editMessageText(chatID, msgID, "Главное меню:", mainMenuKeyboardMarkup())

	case data == cbBackClubs:
		b.setState(cb.From.ID, stateIdle, 0)
		b.showMyClubs(ctx, chatID, cb.From.ID)

	case strings.HasPrefix(data, cbBindSelect+":"):
		b.handleBindSelect(ctx, cb)

	case strings.HasPrefix(data, cbInviteMember+":"):
		clubID, err := strconv.ParseInt(strings.TrimPrefix(data, cbInviteMember+":"), 10, 64)
		if err != nil {
			b.sendText(chatID, "Ошибка: неверный идентификатор клуба.")
			return
		}
		b.setState(cb.From.ID, stateInviteMember, clubID)
		b.sendText(chatID, "Введите Telegram username приглашаемого (например, @john_doe):")

	case strings.HasPrefix(data, cbListMembers+":"):
		clubID, err := strconv.ParseInt(strings.TrimPrefix(data, cbListMembers+":"), 10, 64)
		if err != nil {
			b.sendText(chatID, "Ошибка: неверный идентификатор клуба.")
			return
		}
		b.showClubMembers(ctx, chatID, msgID, clubID, cb.From.ID)

	case strings.HasPrefix(data, cbMemberAction+":"):
		b.handleMemberAction(ctx, cb)

	case strings.HasPrefix(data, cbAcceptInvite+":"):
		clubID, err := strconv.ParseInt(strings.TrimPrefix(data, cbAcceptInvite+":"), 10, 64)
		if err != nil {
			b.sendText(chatID, "Ошибка: неверный идентификатор клуба.")
			return
		}
		b.handleAcceptInvitation(ctx, cb, clubID)

	case strings.HasPrefix(data, cbRejectInvite+":"):
		clubID, err := strconv.ParseInt(strings.TrimPrefix(data, cbRejectInvite+":"), 10, 64)
		if err != nil {
			b.sendText(chatID, "Ошибка: неверный идентификатор клуба.")
			return
		}
		b.handleRejectInvitation(ctx, cb, clubID)

	case strings.HasPrefix(data, cbConfirmEntry+":"):
		b.handleConfirmEntry(ctx, cb)

	case strings.HasPrefix(data, cbAssignAdmin+":"):
		b.handleAssignAdmin(ctx, cb)

	case strings.HasPrefix(data, cbRemoveAdmin+":"):
		b.handleRemoveAdmin(ctx, cb)

	case strings.HasPrefix(data, cbBanMember+":"):
		b.handleBanMember(ctx, cb)

	case strings.HasPrefix(data, cbUnbanMember+":"):
		b.handleUnbanMember(ctx, cb)

	case strings.HasPrefix(data, cbKickMember+":"):
		b.handleKickMember(ctx, cb)

	case strings.HasPrefix(data, cbBackMembers+":"):
		clubID, err := strconv.ParseInt(strings.TrimPrefix(data, cbBackMembers+":"), 10, 64)
		if err != nil {
			b.sendText(chatID, "Ошибка: неверный идентификатор клуба.")
			return
		}
		b.setState(cb.From.ID, stateIdle, 0)
		b.showClubMembers(ctx, chatID, msgID, clubID, cb.From.ID)
	}
}

// handleTextMessage processes regular text messages based on the user's current state.
func (b *Bot) handleTextMessage(ctx context.Context, update tgbotapi.Update) {
	msg := update.Message
	if msg == nil || msg.Text == "" {
		return
	}

	state := b.getState(msg.From.ID)
	if state == nil || state.action == stateIdle {
		b.sendText(msg.Chat.ID, "Используйте кнопки меню для навигации.")
		return
	}

	switch state.action {
	case stateCreateClub:
		b.handleCreateClub(ctx, msg)
	case stateChangeName:
		b.handleChangeName(ctx, msg, state.clubID)
	case stateInviteMember:
		b.handleInviteMember(ctx, msg, state.clubID, msg.Text)
	default:
		b.sendText(msg.Chat.ID, "Пожалуйста, используйте кнопки для продолжения.")
	}
}

// handleCreateClub creates a club with the name provided by the user.
func (b *Bot) handleCreateClub(ctx context.Context, msg *tgbotapi.Message) {
	clubName := strings.TrimSpace(msg.Text)
	if clubName == "" {
		b.sendText(msg.Chat.ID, "Название клуба не может быть пустым. Попробуйте снова:")
		return
	}

	firstName := msg.From.FirstName
	lastName := msg.From.LastName
	nickname := msg.From.UserName
	if nickname == "" {
		nickname = firstName
	}

	club, err := b.svc.CreateClub(ctx, msg.From.ID, firstName, lastName, nickname, clubName)
	if err != nil {
		b.log.Error("failed to create club", "error", err)
		b.sendText(msg.Chat.ID, fmt.Sprintf("Ошибка при создании клуба: %v", err))
		return
	}

	b.setState(msg.From.ID, stateIdle, 0)
	text := fmt.Sprintf("Клуб «%s» создан!\nID: %d", club.Name, club.ID)
	b.sendTextWithKeyboard(msg.Chat.ID, text, clubMenuKeyboardWithMembers(club.ID))
}

// handleChangeName changes the club name with the new name provided by the user.
func (b *Bot) handleChangeName(ctx context.Context, msg *tgbotapi.Message, clubID int64) {
	newName := strings.TrimSpace(msg.Text)
	if newName == "" {
		b.sendText(msg.Chat.ID, "Название клуба не может быть пустым. Попробуйте снова:")
		return
	}

	if err := b.svc.ChangeClubName(ctx, msg.From.ID, clubID, newName); err != nil {
		b.log.Warn("failed to change club name", "error", err)
		b.sendText(msg.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	b.setState(msg.From.ID, stateIdle, 0)
	b.sendTextWithKeyboard(msg.Chat.ID, "Название клуба изменено.", clubMenuKeyboardWithMembers(clubID))
}

// showMyClubs displays the list of clubs where the user is the owner.
func (b *Bot) showMyClubs(ctx context.Context, chatID int64, tgUserID int64) {
	clubs, err := b.svc.GetUserClubs(ctx, tgUserID)
	if err != nil {
		b.log.Error("failed to get user clubs", "error", err)
		b.sendText(chatID, "Ошибка при получении списка клубов.")
		return
	}

	if len(clubs) == 0 {
		b.sendTextWithKeyboard(chatID, "У вас пока нет клубов.", mainMenuKeyboardMarkup())
		return
	}

	items := make([]clubListItem, len(clubs))
	for i, c := range clubs {
		items[i] = clubListItem{ID: c.ID, Name: c.Name}
	}
	b.sendTextWithKeyboard(chatID, "Ваши клубы:", clubListKeyboard(items))
}

// showClubMenu edits the current message to show the club action menu.
func (b *Bot) showClubMenu(chatID int64, msgID int, clubID int64) {
	b.editMessageText(chatID, msgID, "Меню клуба:", clubMenuKeyboardWithMembers(clubID))
}

// showClubInfo displays detailed information about a club.
func (b *Bot) showClubInfo(ctx context.Context, chatID int64, msgID int, clubID int64) {
	club, err := b.svc.GetClubInfo(ctx, clubID)
	if err != nil {
		b.log.Error("failed to get club info", "error", err)
		b.editMessageText(chatID, msgID, "Ошибка при получении информации о клубе.", clubMenuKeyboardWithMembers(clubID))
		return
	}

	tgChat := "не задан"
	if club.TgChatID != nil {
		tgChat = strconv.FormatInt(*club.TgChatID, 10)
	}

	text := fmt.Sprintf(
		"🏠 Клуб: %s\n"+
			"ID: %d\n"+
			"Telegram чат: %s\n"+
			"Создан: %s",
		club.Name, club.ID, tgChat, club.CreatedAt.Format("02.01.2006 15:04:05"),
	)
	b.editMessageText(chatID, msgID, text, clubMenuKeyboardWithMembers(clubID))
}

// handleBind processes the /bind command sent in a Telegram group.
// It binds the group's chat_id to one of the user's clubs.
func (b *Bot) handleBind(ctx context.Context, msg *tgbotapi.Message) {
	// The command must be sent in a group, not a private chat.
	if msg.Chat.Type != "group" && msg.Chat.Type != "supergroup" {
		b.sendText(msg.Chat.ID, "Команда /bind доступна только в Telegram-группе.")
		return
	}

	tgChatID := msg.Chat.ID
	tgUserID := msg.From.ID

	clubs, err := b.svc.GetUserClubs(ctx, tgUserID)
	if err != nil {
		b.log.Error("failed to get user clubs for bind", "error", err)
		b.sendText(msg.Chat.ID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	if len(clubs) == 0 {
		b.sendText(msg.Chat.ID, "Вы не являетесь владельцем ни одного клуба.")
		return
	}

	if len(clubs) == 1 {
		b.bindGroupToClub(ctx, tgUserID, tgChatID, clubs[0].ID, msg.Chat.ID)
		return
	}

	// Multiple clubs — show inline keyboard to select.
	items := make([]clubListItem, len(clubs))
	for i, c := range clubs {
		items[i] = clubListItem{ID: c.ID, Name: c.Name}
	}
	b.sendTextWithKeyboard(msg.Chat.ID, "Выберите клуб для привязки:", bindClubSelectKeyboard(items, tgChatID))
}

// bindGroupToClub calls the service to bind a group to a club and sends the result.
func (b *Bot) bindGroupToClub(ctx context.Context, tgUserID int64, tgChatID int64, clubID int64, chatID int64) {
	club, err := b.svc.BindGroupToClub(ctx, tgUserID, clubID, tgChatID)
	if err != nil {
		b.log.Warn("failed to bind group to club", "error", err)
		b.sendText(chatID, "Не удалось привязать группу. Убедитесь, что вы владелец клуба и группа ещё не привязана к другому клубу.")
		return
	}
	b.sendText(chatID, fmt.Sprintf("Группа успешно привязана к клубу «%s» (ID: %d).", club.Name, club.ID))
}

// handleBindSelect processes the callback when a user selects a club to bind.
func (b *Bot) handleBindSelect(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: bind_select:<club_id>:<chat_id>
	data := cb.Data
	rest := strings.TrimPrefix(data, cbBindSelect+":")
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
	tgChatID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор группы.")
		return
	}

	b.bindGroupToClub(ctx, cb.From.ID, tgChatID, clubID, cb.Message.Chat.ID)
}

// --- Phase 02: Club member management handlers ---

// handleInviteCommand processes the /invite @username command sent in a group.
// It determines the club from the chat's tg_chat_id, parses the username from
// the command arguments, and delegates to handleInviteMember.
func (b *Bot) handleInviteCommand(ctx context.Context, msg *tgbotapi.Message) {
	// The command must be sent in a group, not a private chat.
	if msg.Chat.Type != "group" && msg.Chat.Type != "supergroup" {
		b.sendText(msg.Chat.ID, "Команда /invite доступна только в Telegram-группе.")
		return
	}

	// Determine the club from the group's tg_chat_id.
	club, err := b.svc.GetClubByTgChatID(ctx, msg.Chat.ID)
	if err != nil {
		b.sendText(msg.Chat.ID, "Эта группа не привязана к клубу.")
		return
	}

	// Parse the username from the command arguments.
	username := msg.CommandArguments()
	b.handleInviteMember(ctx, msg, club.ID, username)
}

// handleInviteMember processes the text input containing the Telegram username
// of the user to invite. It verifies that the user is a member of the group
// via getChatMember, then creates a pending club_member and publishes an
// invitation message in the group with a deep link button.
func (b *Bot) handleInviteMember(ctx context.Context, msg *tgbotapi.Message, clubID int64, username string) {
	username = strings.TrimSpace(username)
	username = strings.TrimPrefix(username, "@")
	if username == "" {
		b.sendText(msg.Chat.ID, "Username не может быть пустым. Попробуйте снова:")
		return
	}

	// Look up the player by username (must already be registered in players).
	player, err := b.svc.GetPlayerByUsername(ctx, username)
	if err != nil {
		b.log.Warn("player not found by username", "username", username)
		b.sendTextWithKeyboard(msg.Chat.ID, fmt.Sprintf("Пользователь @%s не найден в системе.", username), clubMenuKeyboardWithMembers(clubID))
		return
	}

	if player.TgUserID == nil {
		b.sendTextWithKeyboard(msg.Chat.ID, fmt.Sprintf("У пользователя @%s не привязан Telegram ID.", username), clubMenuKeyboardWithMembers(clubID))
		return
	}

	// Verify the user is actually a member of this Telegram group via getChatMember.
	chatMember, err := b.api.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: msg.Chat.ID,
			UserID: *player.TgUserID,
		},
	})
	if err != nil {
		b.log.Error("failed to check chat member", "error", err, "tg_user_id", *player.TgUserID)
		b.sendTextWithKeyboard(msg.Chat.ID, "Ошибка при проверке участия пользователя в группе.", clubMenuKeyboardWithMembers(clubID))
		return
	}

	if chatMember.HasLeft() || chatMember.WasKicked() {
		b.sendTextWithKeyboard(msg.Chat.ID, fmt.Sprintf("Пользователь @%s не является участником этой группы.", username), clubMenuKeyboardWithMembers(clubID))
		return
	}

	// Create the club_member with status 'pending' via service.
	_, club, err := b.svc.InviteMember(ctx, msg.From.ID, clubID, *player.TgUserID)
	if err != nil {
		b.log.Warn("failed to invite member", "error", err)
		b.sendTextWithKeyboard(msg.Chat.ID, fmt.Sprintf("Ошибка: %v", err), clubMenuKeyboardWithMembers(clubID))
		return
	}

	b.setState(msg.From.ID, stateIdle, 0)

	// Publish invitation message in the group with a deep link button.
	inviteText := fmt.Sprintf("Приглашение в клуб «%s» для @%s.\nНажмите кнопку, чтобы перейти в личный чат с ботом.", club.Name, username)
	b.sendTextWithKeyboard(msg.Chat.ID, inviteText, groupInviteKeyboard(club.ID, b.api.Self.UserName))

	b.sendTextWithKeyboard(msg.Chat.ID, fmt.Sprintf("Приглашение опубликовано для @%s.", username), clubMenuKeyboardWithMembers(clubID))
}

// showClubMembers displays the list of club members.
func (b *Bot) showClubMembers(ctx context.Context, chatID int64, msgID int, clubID int64, tgUserID int64) {
	members, err := b.svc.GetClubMembers(ctx, tgUserID, clubID)
	if err != nil {
		b.log.Error("failed to get club members", "error", err)
		b.editMessageText(chatID, msgID, "Ошибка при получении списка участников.", clubMenuKeyboardWithMembers(clubID))
		return
	}

	if len(members) == 0 {
		b.editMessageText(chatID, msgID, "В клубе пока нет участников.", clubMenuKeyboardWithMembers(clubID))
		return
	}

	// Build the member list text.
	var sb strings.Builder
	sb.WriteString("Участники клуба:\n\n")
	for i, m := range members {
		name := m.Player.FirstName
		if m.Player.LastName != "" {
			name += " " + m.Player.LastName
		}
		if m.Player.Nickname != "" {
			name += " (@" + m.Player.Nickname + ")"
		}
		sb.WriteString(fmt.Sprintf("%d. %s — %s, %s", i+1, name, roleLabel(m.Role), statusLabel(m.Status)))
		if m.Status == "pending" {
			if m.Accepted {
				sb.WriteString(" (ожидает подтверждения)")
			} else {
				sb.WriteString(" (ожидает ответа)")
			}
		}
		sb.WriteString("\n")
	}

	b.editMessageText(chatID, msgID, sb.String(), memberListKeyboard(clubID, members))
}

// handleMemberAction displays management options for a specific member.
func (b *Bot) handleMemberAction(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: member_action:<club_id>:<player_id>
	parts := strings.SplitN(strings.TrimPrefix(cb.Data, cbMemberAction+":"), ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	playerID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игрока.")
		return
	}

	// Get the member with player info.
	members, err := b.svc.GetClubMembers(ctx, cb.From.ID, clubID)
	if err != nil {
		b.log.Error("failed to get club members", "error", err)
		b.sendText(cb.Message.Chat.ID, "Ошибка при получении данных участника.")
		return
	}

	var targetMember *domain.ClubMemberWithPlayer
	for _, m := range members {
		if m.PlayerID == playerID {
			targetMember = m
			break
		}
	}
	if targetMember == nil {
		b.sendText(cb.Message.Chat.ID, "Участник не найден.")
		return
	}

	// Determine the current user's role in this club.
	userRole := ""
	for _, m := range members {
		if m.Player.TgUserID != nil && *m.Player.TgUserID == cb.From.ID {
			userRole = m.Role
			break
		}
	}

	// Build member info text.
	name := targetMember.Player.FirstName
	if targetMember.Player.LastName != "" {
		name += " " + targetMember.Player.LastName
	}
	if targetMember.Player.Nickname != "" {
		name += " (@" + targetMember.Player.Nickname + ")"
	}

	text := fmt.Sprintf(
		"👤 %s\n"+
			"Роль: %s\n"+
			"Статус: %s",
		name, roleLabel(targetMember.Role), statusLabel(targetMember.Status),
	)
	if targetMember.Status == "pending" {
		if targetMember.Accepted {
			text += "\nПринял приглашение: да"
		} else {
			text += "\nПринял приглашение: нет"
		}
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, text, memberActionKeyboard(clubID, playerID, targetMember, userRole))
}

// handleAcceptInvitation processes the user's acceptance of an invitation.
func (b *Bot) handleAcceptInvitation(ctx context.Context, cb *tgbotapi.CallbackQuery, clubID int64) {
	player, club, notifyIDs, err := b.svc.AcceptInvitation(ctx, cb.From.ID, clubID)
	if err != nil {
		b.log.Warn("failed to accept invitation", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Notify the user that their invitation was accepted.
	b.sendText(cb.Message.Chat.ID, "Вы приняли приглашение в клуб «"+club.Name+"». Ожидайте подтверждения от владельца или администратора.")

	// Notify owner and admin with a confirm button.
	for _, id := range notifyIDs {
		if id == cb.From.ID {
			continue // don't notify the user themselves
		}
		notifyText := fmt.Sprintf(
			"Пользователь %s принял приглашение в клуб «%s» (ID: %d). Подтвердить вступление?",
			cb.From.FirstName, club.Name, club.ID,
		)
		b.sendTextWithKeyboard(id, notifyText, confirmEntryKeyboard(clubID, player.ID))
	}
}

// handleRejectInvitation processes the user's rejection of an invitation.
func (b *Bot) handleRejectInvitation(ctx context.Context, cb *tgbotapi.CallbackQuery, clubID int64) {
	club, notifyIDs, err := b.svc.RejectInvitation(ctx, cb.From.ID, clubID)
	if err != nil {
		b.log.Warn("failed to reject invitation", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Notify the user.
	b.sendText(cb.Message.Chat.ID, "Вы отклонили приглашение в клуб «"+club.Name+"».")

	// Notify owner and admin.
	for _, id := range notifyIDs {
		if id == cb.From.ID {
			continue
		}
		notifyText := fmt.Sprintf("Пользователь %s отклонил приглашение в клуб «%s» (ID: %d).", cb.From.FirstName, club.Name, club.ID)
		b.sendText(id, notifyText)
	}
}

// handleConfirmEntry processes the owner/admin's confirmation of a user's entry.
func (b *Bot) handleConfirmEntry(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: confirm_entry:<club_id>:<player_id>
	parts := strings.SplitN(strings.TrimPrefix(cb.Data, cbConfirmEntry+":"), ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	playerID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игрока.")
		return
	}

	player, club, err := b.svc.ConfirmEntry(ctx, cb.From.ID, clubID, playerID)
	if err != nil {
		b.log.Warn("failed to confirm entry", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Notify the user.
	if player.TgUserID != nil {
		notifyText := fmt.Sprintf("Ваше вступление в клуб «%s» подтверждено. Добро пожаловать!", club.Name)
		b.sendText(*player.TgUserID, notifyText)
	}

	// Update the message.
	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Вступление подтверждено.", clubMenuKeyboardWithMembers(clubID))
}

// handleAssignAdmin processes the owner's assignment of admin role to a member.
func (b *Bot) handleAssignAdmin(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: assign_admin:<club_id>:<player_id>
	parts := strings.SplitN(strings.TrimPrefix(cb.Data, cbAssignAdmin+":"), ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	playerID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игрока.")
		return
	}

	player, err := b.svc.AssignAdmin(ctx, cb.From.ID, clubID, playerID)
	if err != nil {
		b.log.Warn("failed to assign admin", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Notify the user.
	if player.TgUserID != nil {
		b.sendText(*player.TgUserID, "Вас назначили администратором клуба.")
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Администратор назначен.", clubMenuKeyboardWithMembers(clubID))
}

// handleRemoveAdmin processes the owner's removal of admin role from a member.
func (b *Bot) handleRemoveAdmin(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: remove_admin:<club_id>:<player_id>
	parts := strings.SplitN(strings.TrimPrefix(cb.Data, cbRemoveAdmin+":"), ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	playerID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игрока.")
		return
	}

	player, err := b.svc.RemoveAdmin(ctx, cb.From.ID, clubID, playerID)
	if err != nil {
		b.log.Warn("failed to remove admin", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Notify the user.
	if player.TgUserID != nil {
		b.sendText(*player.TgUserID, "Вас сняли с должности администратора клуба.")
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Администратор снят.", clubMenuKeyboardWithMembers(clubID))
}

// handleBanMember processes banning a member.
func (b *Bot) handleBanMember(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: ban_member:<club_id>:<player_id>
	parts := strings.SplitN(strings.TrimPrefix(cb.Data, cbBanMember+":"), ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	playerID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игрока.")
		return
	}

	player, err := b.svc.ChangeMemberStatus(ctx, cb.From.ID, clubID, playerID, "banned")
	if err != nil {
		b.log.Warn("failed to ban member", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Notify the user.
	if player.TgUserID != nil {
		b.sendText(*player.TgUserID, "Вы забанены в клубе.")
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Участник забанен.", clubMenuKeyboardWithMembers(clubID))
}

// handleUnbanMember processes unbanning a member.
func (b *Bot) handleUnbanMember(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: unban_member:<club_id>:<player_id>
	parts := strings.SplitN(strings.TrimPrefix(cb.Data, cbUnbanMember+":"), ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	playerID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игрока.")
		return
	}

	player, err := b.svc.ChangeMemberStatus(ctx, cb.From.ID, clubID, playerID, "active")
	if err != nil {
		b.log.Warn("failed to unban member", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Notify the user.
	if player.TgUserID != nil {
		b.sendText(*player.TgUserID, "Вы разбанены в клубе.")
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Участник разбанен.", clubMenuKeyboardWithMembers(clubID))
}

// handleKickMember processes removing (kicking) a member from the club.
func (b *Bot) handleKickMember(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	// Callback data format: kick_member:<club_id>:<player_id>
	parts := strings.SplitN(strings.TrimPrefix(cb.Data, cbKickMember+":"), ":", 2)
	if len(parts) != 2 {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверные данные.")
		return
	}
	clubID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор клуба.")
		return
	}
	playerID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		b.sendText(cb.Message.Chat.ID, "Ошибка: неверный идентификатор игрока.")
		return
	}

	player, err := b.svc.RemoveMember(ctx, cb.From.ID, clubID, playerID)
	if err != nil {
		b.log.Warn("failed to remove member", "error", err)
		b.sendText(cb.Message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Notify the user.
	if player.TgUserID != nil {
		b.sendText(*player.TgUserID, "Вы удалены из клуба.")
	}

	b.editMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Участник исключен.", clubMenuKeyboardWithMembers(clubID))
}
