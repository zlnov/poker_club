package telegram

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"poker-club/backend/internal/domain"
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

	// Phase 02: club member management
	cbInviteMember = "invite_member"
	cbListMembers  = "list_members"
	cbMemberAction = "member_action"
	cbAcceptInvite = "accept_invite"
	cbRejectInvite = "reject_invite"
	cbConfirmEntry = "confirm_entry"
	cbAssignAdmin  = "assign_admin"
	cbRemoveAdmin  = "remove_admin"
	cbBanMember    = "ban_member"
	cbUnbanMember  = "unban_member"
	cbKickMember   = "kick_member"
	cbBackMembers  = "back_members"

	// Club menu structure
	cbClubMenu       = "club_menu"
	cbManageMenu     = "manage_menu"
	cbBackToClubMenu = "back_to_club_menu"
)

// stateAction constants for user input state
const (
	stateIdle         = ""
	stateCreateClub   = "create_club"
	stateChangeName   = "change_name"
	stateCloseConfirm = "close_confirm"
	stateInviteMember = "invite_member"
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

// clubMainMenuKeyboard returns the intermediate club menu with "Клуб" and
// "Управление" buttons. "Управление" is only included for owner/admin roles.
func clubMainMenuKeyboard(clubID int64, userRole string, isPrivate bool) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 3)

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Клуб", fmt.Sprintf("%s:%s", cbClubMenu, id)),
	))

	if userRole == "owner" || userRole == "admin" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Управление", fmt.Sprintf("%s:%s", cbManageMenu, id)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", cbBackClubs),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// clubSubMenuKeyboard returns the "Клуб" submenu with info and member list.
// "Закрыть клуб" is only included for the owner in a private chat.
func clubSubMenuKeyboard(clubID int64, userRole string, isPrivate bool) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 4)

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Инфо", fmt.Sprintf("%s:%s", cbClubInfo, id)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Список участников", fmt.Sprintf("%s:%s", cbListMembers, id)),
	))

	// if userRole == "owner" && isPrivate {
	// 	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
	// 		tgbotapi.NewInlineKeyboardButtonData("Закрыть клуб", fmt.Sprintf("%s:%s", cbCloseClub, id)),
	// 	))
	// }

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s", cbBackToClubMenu, id)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// manageSubMenuKeyboard returns the "Управление" submenu with club management actions.
func manageSubMenuKeyboard(clubID int64, userRole string) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)

	// return tgbotapi.NewInlineKeyboardMarkup(
	// 	tgbotapi.NewInlineKeyboardRow(
	// 		tgbotapi.NewInlineKeyboardButtonData("Изменить название", fmt.Sprintf("%s:%s", cbChangeName, id)),
	// 	),
	// 	tgbotapi.NewInlineKeyboardRow(
	// 		tgbotapi.NewInlineKeyboardButtonData("Пригласить участника", fmt.Sprintf("%s:%s", cbInviteMember, id)),
	// 	),
	// 	tgbotapi.NewInlineKeyboardRow(
	// 		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s", cbBackToClubMenu, id)),
	// 	),
	// )

	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 4)

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Изменить название", fmt.Sprintf("%s:%s", cbChangeName, id)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Пригласить участника", fmt.Sprintf("%s:%s", cbInviteMember, id)),
	))

	if userRole == "owner" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Закрыть клуб", fmt.Sprintf("%s:%s", cbCloseClub, id)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s", cbBackToClubMenu, id)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}

}

// memberListKeyboard builds an inline keyboard listing club members for selection.
func memberListKeyboard(clubID int64, members []*domain.ClubMemberWithPlayer) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(members)+1)
	for _, m := range members {
		label := memberLabel(m)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%d:%d", cbMemberAction, clubID, m.PlayerID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", cbBackClubs),
	))
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// memberLabel returns a short label for a member in the list.
func memberLabel(m *domain.ClubMemberWithPlayer) string {
	name := m.Player.FirstName
	if m.Player.LastName != "" {
		name += " " + m.Player.LastName
	}
	if m.Player.Nickname != "" && m.Player.Nickname != m.Player.FirstName {
		name += " (@" + m.Player.Nickname + ")"
	}
	return fmt.Sprintf("%s [%s, %s]", name, roleLabel(m.Role), statusLabel(m.Status))
}

func roleLabel(role string) string {
	switch role {
	case "owner":
		return "владелец"
	case "admin":
		return "админ"
	case "member":
		return "участник"
	default:
		return role
	}
}

func statusLabel(status string) string {
	switch status {
	case "pending":
		return "pending"
	case "active":
		return "active"
	case "banned":
		return "banned"
	case "left":
		return "left"
	default:
		return status
	}
}

// memberActionKeyboard builds an inline keyboard for managing a specific member.
// userRole determines which action buttons are shown (owner sees all, admin sees
// management actions, member sees none).
func memberActionKeyboard(clubID, playerID int64, member *domain.ClubMemberWithPlayer, userRole string) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	pid := strconv.FormatInt(playerID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 6)

	canManage := userRole == "owner" || userRole == "admin"

	// Confirm entry (only if pending and accepted, owner/admin only)
	if canManage && member.Status == "pending" && member.Accepted {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить вступление", fmt.Sprintf("%s:%s:%s", cbConfirmEntry, id, pid)),
		))
	}

	// Assign/Remove admin (owner only)
	if userRole == "owner" {
		if member.Role == "member" {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назначить админом", fmt.Sprintf("%s:%s:%s", cbAssignAdmin, id, pid)),
			))
		} else if member.Role == "admin" {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Снять админа", fmt.Sprintf("%s:%s:%s", cbRemoveAdmin, id, pid)),
			))
		}
	}

	// Ban/Unban (owner/admin)
	if canManage {
		if member.Status == "active" {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Забанить", fmt.Sprintf("%s:%s:%s", cbBanMember, id, pid)),
			))
		} else if member.Status == "banned" {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Разбанить", fmt.Sprintf("%s:%s:%s", cbUnbanMember, id, pid)),
			))
		}
	}

	// Kick (change to left) - only for active or banned members (owner/admin)
	if canManage && (member.Status == "active" || member.Status == "banned") {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Исключить", fmt.Sprintf("%s:%s:%s", cbKickMember, id, pid)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад к списку", fmt.Sprintf("%s:%s", cbBackMembers, id)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// invitationKeyboard returns the Accept/Reject inline keyboard for an invitation message.
func invitationKeyboard(clubID int64) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Принять", fmt.Sprintf("%s:%s", cbAcceptInvite, id)),
			tgbotapi.NewInlineKeyboardButtonData("Отклонить", fmt.Sprintf("%s:%s", cbRejectInvite, id)),
		),
	)
}

// groupInviteKeyboard returns an inline keyboard with a deep link button that
// opens a private chat with the bot. The deep link parameter encodes the club ID
// so the bot can identify which invitation to show.
func groupInviteKeyboard(clubID int64, botUsername string) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	deepLink := fmt.Sprintf("https://t.me/%s?start=invite_%s", botUsername, id)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Принять приглашение", deepLink),
		),
	)
}

// confirmEntryKeyboard returns the Confirm inline keyboard for a notification to owner/admin.
func confirmEntryKeyboard(clubID, playerID int64) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	pid := strconv.FormatInt(playerID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить вступление", fmt.Sprintf("%s:%s:%s", cbConfirmEntry, id, pid)),
		),
	)
}
