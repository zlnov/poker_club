package telegram

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// callback data constants
const (
	cbCreateClub   = "create_club"
	cbMyClubs      = "my_clubs"
	cbSelectClub   = "select_club"
	cbClubInfo     = "club_info"
	cbChangeName   = "change_name"
	cbCloseClub    = "close_club"
	cbConfirmClose = "confirm_close"
	cbCancelClose  = "cancel_close"
	cbBackMain     = "back_main"
	cbBackClubs    = "back_clubs"
	cbBindSelect   = "bind_select"
)

// stateAction constants for user input state
const (
	stateIdle         = ""
	stateCreateClub   = "create_club"
	stateChangeName   = "change_name"
	stateCloseConfirm = "close_confirm"
)

// mainMenuKeyboardMarkup returns the inline keyboard for the main menu.
func mainMenuKeyboardMarkup() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Создать клуб", cbCreateClub),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Мои клубы", cbMyClubs),
		),
	)
}

// clubListKeyboard builds an inline keyboard listing the user's clubs.
func clubListKeyboard(clubs []clubListItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(clubs)+1)
	for _, c := range clubs {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(c.Name, fmt.Sprintf("%s:%d", cbSelectClub, c.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", cbBackMain),
	))
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// clubMenuKeyboard returns the inline keyboard for a club's action menu.
func clubMenuKeyboard(clubID int64) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Инфо", fmt.Sprintf("%s:%s", cbClubInfo, id)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Изменить название", fmt.Sprintf("%s:%s", cbChangeName, id)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Закрыть клуб", fmt.Sprintf("%s:%s", cbCloseClub, id)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", cbBackMain),
		),
	)
}

// confirmCloseKeyboard returns a Yes/No inline keyboard for club deletion confirmation.
func confirmCloseKeyboard(clubID int64) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", fmt.Sprintf("%s:%s", cbConfirmClose, id)),
			tgbotapi.NewInlineKeyboardButtonData("Нет", fmt.Sprintf("%s:%s", cbCancelClose, id)),
		),
	)
}

// clubListItem is a lightweight representation of a club for list display.
type clubListItem struct {
	ID   int64
	Name string
}

// bindClubSelectKeyboard builds an inline keyboard for selecting a club to bind a group to.
func bindClubSelectKeyboard(clubs []clubListItem, tgChatID int64) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(clubs))
	for _, c := range clubs {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(c.Name, fmt.Sprintf("%s:%d:%d", cbBindSelect, c.ID, tgChatID)),
		))
	}
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}
