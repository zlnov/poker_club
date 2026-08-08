package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	default:
		b.sendText(msg.Chat.ID, "Неизвестная команда. Используйте кнопки меню.")
	}
}

// handleStart processes the /start command — shows the main menu.
func (b *Bot) handleStart(ctx context.Context, msg *tgbotapi.Message) {
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
	b.sendTextWithKeyboard(msg.Chat.ID, text, clubMenuKeyboard(club.ID))
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
	b.sendTextWithKeyboard(msg.Chat.ID, "Название клуба изменено.", clubMenuKeyboard(clubID))
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
	b.editMessageText(chatID, msgID, "Меню клуба:", clubMenuKeyboard(clubID))
}

// showClubInfo displays detailed information about a club.
func (b *Bot) showClubInfo(ctx context.Context, chatID int64, msgID int, clubID int64) {
	club, err := b.svc.GetClubInfo(ctx, clubID)
	if err != nil {
		b.log.Error("failed to get club info", "error", err)
		b.editMessageText(chatID, msgID, "Ошибка при получении информации о клубе.", clubMenuKeyboard(clubID))
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
	b.editMessageText(chatID, msgID, text, clubMenuKeyboard(clubID))
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
