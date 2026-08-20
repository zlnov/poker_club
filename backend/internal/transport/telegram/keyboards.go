package telegram

import (
	"fmt"
	"strconv"
	"time"

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
	cbMenu         = "menu"

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

	// Phase 03: game management
	cbCreateGame               = "create_game"
	cbGameList                 = "game_list"
	cbSelectGame               = "select_game"
	cbGameInfo                 = "game_info"
	cbGameParticipants         = "game_participants"
	cbGameInvite               = "game_invite"
	cbGameChangeParams         = "game_change_params"
	cbGameCancel               = "game_cancel"
	cbGameConfirmCancel        = "game_confirm_cancel"
	cbGameAccept               = "game_accept"
	cbGameDecline              = "game_decline"
	cbGameConfirmParticipation = "game_confirm_participation"
	cbGameRemoveParticipant    = "game_remove_participant"
	cbGameBack                 = "game_back"
	cbGameEditParam            = "game_edit_param"
	cbGameSelectParam          = "game_select_param"
	cbGameCreateConfirm        = "game_create_confirm"
	cbGameCreateCancel         = "game_create_cancel"

	// Phase 04: active game management
	cbGameStart            = "game_start"
	cbGameCompleteBuyin    = "game_complete_buyin"
	cbGameMonitor          = "game_monitor"
	cbGamePlayerMonitor    = "game_player_monitor"
	cbGameBank             = "game_bank"
	cbGameExpenses         = "game_expenses"
	cbGameRebuy            = "game_rebuy"
	cbGameRebuyPlayer      = "game_rebuy_player"
	cbGameRebuyConfirm     = "game_rebuy_confirm"
	cbGameRebuyManage      = "game_rebuy_manage"
	cbGameRebuyFixOp       = "game_rebuy_fix_op"
	cbGameRebuyFixConfirm  = "game_rebuy_fix_confirm"
	cbGameAddPlayer        = "game_add_player"
	cbGameAddPlayerConfirm = "game_add_player_confirm"
	cbGamePauseTimer       = "game_pause_timer"
	cbGameResumeTimer      = "game_resume_timer"
	cbGameExtend           = "game_extend"
	cbGameExtendSelect     = "game_extend_select"
	cbGamePlayerStack      = "game_player_stack"
	cbGamePlayerStats      = "game_player_stats"
	cbGameStats            = "game_stats"
	cbGameActiveBack       = "game_active_back"

	// Phase 05: game end
	cbGameEnd           = "game_end"
	cbGameEndPlayer     = "game_end_player"
	cbGameEndCheckBank  = "game_end_check_bank"
	cbGameEndConfirm    = "game_end_confirm"
	cbGameEndFinish     = "game_end_finish"
)

// stateAction constants for user input state
const (
	stateIdle         = ""
	stateCreateClub   = "create_club"
	stateChangeName   = "change_name"
	stateCloseConfirm = "close_confirm"
	stateInviteMember = "invite_member"

	// Phase 03: game creation states
	stateCreateGame       = "create_game"
	stateGameEditParam    = "game_edit_param"
	stateGameInviteMember = "game_invite_member"

	// Phase 04: active game states
	stateGamePlayerStack   = "game_player_stack"
	stateGameRebuyFixCount = "game_rebuy_fix_count"

	// Phase 05: game end states
	stateGameEndChipsInput = "game_end_chips_input"
)

// mainMenuKeyboardMarkup returns the inline keyboard for the main menu.
// This is the Owner/Admin version with "Создать клуб".
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

// privateMemberMenuKeyboard returns the inline keyboard for the main menu
// in a private chat for a Member (no "Создать клуб" button).
func privateMemberMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Мои клубы", cbMyClubs),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", cbBackMain),
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
// "Управление" buttons. In group chat, "Управление" is a URL button with a
// deep link to private chat for all roles. In private chat, "Управление" is
// a callback button only for owner/admin.
func clubMainMenuKeyboard(clubID int64, userRole string, isPrivate bool, botUsername string) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 3)

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Клуб", fmt.Sprintf("%s:%s", cbClubMenu, id)),
	))

	if isPrivate {
		// Private chat: show "Управление" as callback button for owner/admin only.
		if userRole == "owner" || userRole == "admin" {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Управление", fmt.Sprintf("%s:%s", cbManageMenu, id)),
			))
		}
	} else {
		// Group chat: show "Управление" as URL button (deep link) for all roles.
		deepLink := fmt.Sprintf("https://t.me/%s?start=club_%s", botUsername, id)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Управление", deepLink),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", cbBackClubs),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// clubSubMenuKeyboard returns the "Клуб" submenu with info and member list.
func clubSubMenuKeyboard(clubID int64, userRole string, isPrivate bool) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 4)

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Инфо", fmt.Sprintf("%s:%s", cbClubInfo, id)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Список участников", fmt.Sprintf("%s:%s", cbListMembers, id)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s", cbBackToClubMenu, id)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// manageSubMenuKeyboard returns the "Управление" submenu with club management actions.
func manageSubMenuKeyboard(clubID int64, userRole string) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)

	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 7)

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Изменить название", fmt.Sprintf("%s:%s", cbChangeName, id)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Создать игру", fmt.Sprintf("%s:%s", cbCreateGame, id)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Список игр", fmt.Sprintf("%s:%s", cbGameList, id)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Состав клуба", fmt.Sprintf("%s:%s", cbListMembers, id)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Пригласить участников", fmt.Sprintf("%s:%s", cbInviteMember, id)),
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

// --- Phase 03: Game management keyboards ---

// gameCreationData stores intermediate game creation parameters.
type gameCreationData struct {
	gameType       string // cash, tournament
	bankerID       int64  // club_members.id
	bankerName     string
	currency       string
	moneyModel     string
	chipValue      string
	buyInAmount    string
	rebuyAllowed   bool
	rebuyPrice     string
	maxRebuys      string
	duration       string
	startTime      string
	minPlayers     string
	maxPlayers     string
	rankingPrimary string
	editingParam   string
}

// gameCreationKeyboard returns the interactive game creation screen.
func gameCreationKeyboard(data *gameCreationData, clubID int64, bankerOptions []clubListItem) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 20)

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Тип игры: "+gameTypeLabel(data.gameType), fmt.Sprintf("%s:%s:type", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Банкир: "+data.bankerName, fmt.Sprintf("%s:%s:banker", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Валюта: "+data.currency, fmt.Sprintf("%s:%s:currency", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Денежная модель: "+data.moneyModel, fmt.Sprintf("%s:%s:money_model", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Chip value: "+data.chipValue, fmt.Sprintf("%s:%s:chip_value", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Buy-in: "+data.buyInAmount, fmt.Sprintf("%s:%s:buy_in", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Rebuy: "+boolYesNo(data.rebuyAllowed), fmt.Sprintf("%s:%s:rebuy_allowed", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Rebuy price: "+data.rebuyPrice, fmt.Sprintf("%s:%s:rebuy_price", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Max rebuys: "+data.maxRebuys, fmt.Sprintf("%s:%s:max_rebuys", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Start: "+data.startTime, fmt.Sprintf("%s:%s:start_time", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Duration: "+data.duration, fmt.Sprintf("%s:%s:duration", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Min players: "+data.minPlayers, fmt.Sprintf("%s:%s:min_players", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Max players: "+data.maxPlayers, fmt.Sprintf("%s:%s:max_players", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Ranking: "+data.rankingPrimary, fmt.Sprintf("%s:%s:ranking_primary", cbGameEditParam, id)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Создать", fmt.Sprintf("%s:%s", cbGameCreateConfirm, id)),
		tgbotapi.NewInlineKeyboardButtonData("Отмена", fmt.Sprintf("%s:%s", cbGameCreateCancel, id)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// gameTypeKeyboard returns the inline keyboard for selecting game type.
func gameTypeKeyboard(clubID int64) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Cash", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "type:cash")),
			tgbotapi.NewInlineKeyboardButtonData("Tournament", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "type:tournament")),
		),
	)
}

// bankerSelectKeyboard returns the inline keyboard for selecting a banker from club members.
func bankerSelectKeyboard(clubID int64, members []*domain.ClubMemberWithPlayer) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(members)+1)
	for _, m := range members {
		label := m.Player.FirstName
		if m.Player.LastName != "" {
			label += " " + m.Player.LastName
		}
		if m.Player.Nickname != "" && m.Player.Nickname != m.Player.FirstName {
			label += " (@" + m.Player.Nickname + ")"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%s:%s:%d", cbGameSelectParam, id, "banker", m.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Отмена", fmt.Sprintf("%s:%s", cbGameCreateCancel, id)),
	))
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// currencyKeyboard returns the inline keyboard for selecting currency.
func currencyKeyboard(clubID int64) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("RUB", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "currency:RUB")),
			tgbotapi.NewInlineKeyboardButtonData("USD", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "currency:USD")),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("EUR", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "currency:EUR")),
			tgbotapi.NewInlineKeyboardButtonData("KZT", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "currency:KZT")),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("BYN", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "currency:BYN")),
		),
	)
}

// moneyModelKeyboard returns the inline keyboard for selecting money model.
func moneyModelKeyboard(clubID int64) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("real", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "money_model:real")),
			tgbotapi.NewInlineKeyboardButtonData("points", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "money_model:points")),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("virtual", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "money_model:virtual")),
			tgbotapi.NewInlineKeyboardButtonData("practice", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "money_model:practice")),
		),
	)
}

// rebuyAllowedKeyboard returns the inline keyboard for selecting rebuy allowed.
func rebuyAllowedKeyboard(clubID int64) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "rebuy_allowed:yes")),
			tgbotapi.NewInlineKeyboardButtonData("Нет", fmt.Sprintf("%s:%s:%s", cbGameSelectParam, id, "rebuy_allowed:no")),
		),
	)
}

// gameListKeyboard returns the inline keyboard for listing games.
func gameListKeyboard(clubID int64, games []*domain.Game) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(clubID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(games)+2)
	for _, g := range games {
		label := fmt.Sprintf("%s [%s]", gameTypeLabel(g.GameType), g.Status)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%s:%d", cbSelectGame, id, g.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s", cbBackToClubMenu, id)),
	))
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// gameMenuKeyboard returns the inline keyboard for the game menu.
// For planned games, shows "Начать игру" for banker/owner/admin.
func gameMenuKeyboard(clubID, gameID int64, userRole string, isBanker bool, gameStatus string) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 8)

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Инфо", fmt.Sprintf("%s:%s:%s", cbGameInfo, cid, gid)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Участники", fmt.Sprintf("%s:%s:%s", cbGameParticipants, cid, gid)),
	))

	canManage := isBanker || userRole == "owner" || userRole == "admin"

	if gameStatus == "planned" && canManage {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Начать игру", fmt.Sprintf("%s:%s:%s", cbGameStart, cid, gid)),
		))
	}

	if canManage {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пригласить", fmt.Sprintf("%s:%s:%s", cbGameInvite, cid, gid)),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Изменить параметры", fmt.Sprintf("%s:%s:%s", cbGameChangeParams, cid, gid)),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Отменить игру", fmt.Sprintf("%s:%s:%s", cbGameCancel, cid, gid)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s", cbBackToClubMenu, cid)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// gameParticipantListKeyboard returns the inline keyboard for listing game participants.
func gameParticipantListKeyboard(clubID, gameID int64, participants []*domain.GameParticipantWithPlayer, userRole string) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(participants)+2)

	for _, p := range participants {
		label := p.Player.FirstName
		if p.Player.LastName != "" {
			label += " " + p.Player.LastName
		}
		if p.Player.Nickname != "" && p.Player.Nickname != p.Player.FirstName {
			label += " (@" + p.Player.Nickname + ")"
		}
		label += " [" + participantStatusLabel(p.Status) + "]"
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%s:%s:%d", cbMemberAction, cid, gid, p.PlayerID)),
		))
	}

	if userRole == "owner" || userRole == "admin" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пригласить участника", fmt.Sprintf("%s:%s:%s", cbGameInvite, cid, gid)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameBack, cid, gid)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// gameParticipantActionKeyboard returns the inline keyboard for managing a specific game participant.
func gameParticipantActionKeyboard(clubID, gameID, playerID int64, participant *domain.GameParticipantWithPlayer, userRole string) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	pid := strconv.FormatInt(playerID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 4)

	canManage := userRole == "owner" || userRole == "admin"

	if canManage && participant.Status == "accepted" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить участие", fmt.Sprintf("%s:%s:%s:%s", cbGameConfirmParticipation, cid, gid, pid)),
		))
	}

	if canManage && (participant.Status == "invited" || participant.Status == "accepted" || participant.Status == "confirmed") {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Удалить", fmt.Sprintf("%s:%s:%s:%s", cbGameRemoveParticipant, cid, gid, pid)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад к списку", fmt.Sprintf("%s:%s:%s", cbGameParticipants, cid, gid)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// gameConfirmCancelKeyboard returns the confirm/cancel keyboard for game cancellation.
func gameConfirmCancelKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", fmt.Sprintf("%s:%s:%s", cbGameConfirmCancel, cid, gid)),
			tgbotapi.NewInlineKeyboardButtonData("Нет", fmt.Sprintf("%s:%s:%s", cbGameBack, cid, gid)),
		),
	)
}

// gameAcceptDeclineKeyboard returns the Accept/Decline keyboard for game participation.
func gameAcceptDeclineKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Принять", fmt.Sprintf("%s:%s:%s", cbGameAccept, cid, gid)),
			tgbotapi.NewInlineKeyboardButtonData("Отказаться", fmt.Sprintf("%s:%s:%s", cbGameDecline, cid, gid)),
		),
	)
}

// gameConfirmParticipationKeyboard returns the Confirm/Decline keyboard for owner/admin.
func gameConfirmParticipationKeyboard(clubID, gameID, playerID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	pid := strconv.FormatInt(playerID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить", fmt.Sprintf("%s:%s:%s:%s", cbGameConfirmParticipation, cid, gid, pid)),
		),
	)
}

// gameInviteMemberKeyboard returns the keyboard for inviting a member to a game.
func gameInviteMemberKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameBack, cid, gid)),
		),
	)
}

// gameTypeLabel returns a human-readable label for a game type.
func gameTypeLabel(gt string) string {
	switch gt {
	case "cash":
		return "Cash"
	case "tournament":
		return "Tournament"
	default:
		return gt
	}
}

// participantStatusLabel returns a human-readable label for a game participant status.
func participantStatusLabel(status string) string {
	switch status {
	case "invited":
		return "invited"
	case "accepted":
		return "accepted"
	case "declined":
		return "declined"
	case "confirmed":
		return "confirmed"
	default:
		return status
	}
}

// boolYesNo returns "Да" or "Нет" for a boolean value.
func boolYesNo(b bool) string {
	if b {
		return "Да"
	}
	return "Нет"
}

// --- Phase 04: Active game keyboards ---

// activeGameMenuKeyboard returns the inline keyboard for an active game.
// Shows different options based on user role and whether they are the banker.
func activeGameMenuKeyboard(clubID, gameID int64, userRole string, isBanker bool, gameType string, hasTimer bool) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 12)

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Монитор игры", fmt.Sprintf("%s:%s:%s", cbGameMonitor, cid, gid)),
	))

	canManage := isBanker || userRole == "owner" || userRole == "admin"

	if canManage {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Текущий банк", fmt.Sprintf("%s:%s:%s", cbGameBank, cid, gid)),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Расходы игроков", fmt.Sprintf("%s:%s:%s", cbGameExpenses, cid, gid)),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Rebuy", fmt.Sprintf("%s:%s:%s", cbGameRebuy, cid, gid)),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Управление Rebuy", fmt.Sprintf("%s:%s:%s", cbGameRebuyManage, cid, gid)),
		))
		if gameType == "cash" {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Добавить игрока", fmt.Sprintf("%s:%s:%s", cbGameAddPlayer, cid, gid)),
			))
		}
		if hasTimer {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Пауза", fmt.Sprintf("%s:%s:%s", cbGamePauseTimer, cid, gid)),
			))
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Статистика игры", fmt.Sprintf("%s:%s:%s", cbGameStats, cid, gid)),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Завершить игру", fmt.Sprintf("%s:%s:%s", cbGameEnd, cid, gid)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Моя статистика", fmt.Sprintf("%s:%s:%s", cbGamePlayerStats, cid, gid)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Ввести текущий стек", fmt.Sprintf("%s:%s:%s", cbGamePlayerStack, cid, gid)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameActiveBack, cid, gid)),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// gameMonitorKeyboard returns the keyboard for the game monitor view.
func gameMonitorKeyboard(clubID, gameID int64, canManage bool, gameType string, hasTimer bool) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 8)

	if canManage {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Текущий банк", fmt.Sprintf("%s:%s:%s", cbGameBank, cid, gid)),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Расходы игроков", fmt.Sprintf("%s:%s:%s", cbGameExpenses, cid, gid)),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Rebuy", fmt.Sprintf("%s:%s:%s", cbGameRebuy, cid, gid)),
		))
		if gameType == "cash" {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Добавить игрока", fmt.Sprintf("%s:%s:%s", cbGameAddPlayer, cid, gid)),
			))
		}
		if hasTimer {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Пауза", fmt.Sprintf("%s:%s:%s", cbGamePauseTimer, cid, gid)),
			))
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Статистика игры", fmt.Sprintf("%s:%s:%s", cbGameStats, cid, gid)),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Завершить игру", fmt.Sprintf("%s:%s:%s", cbGameEnd, cid, gid)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Моя статистика", fmt.Sprintf("%s:%s:%s", cbGamePlayerStats, cid, gid)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Ввести текущий стек", fmt.Sprintf("%s:%s:%s", cbGamePlayerStack, cid, gid)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameActiveBack, cid, gid)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// gameCompleteBuyinKeyboard returns the keyboard after game start with buy-in registered.
func gameCompleteBuyinKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Готово", fmt.Sprintf("%s:%s:%s", cbGameCompleteBuyin, cid, gid)),
		),
	)
}

// rebuyPlayerSelectKeyboard builds an inline keyboard for selecting a player to rebuy.
func rebuyPlayerSelectKeyboard(clubID, gameID int64, participants []*domain.GameParticipantWithPlayer) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(participants)+1)

	for _, p := range participants {
		label := p.Player.FirstName
		if p.Player.LastName != "" {
			label += " " + p.Player.LastName
		}
		if p.Player.Nickname != "" && p.Player.Nickname != p.Player.FirstName {
			label += " (@" + p.Player.Nickname + ")"
		}
		label += fmt.Sprintf(" [Rebuy: %d]", p.RebuyCount)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%s:%s:%d", cbGameRebuyPlayer, cid, gid, p.PlayerID)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameActiveBack, cid, gid)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// rebuyConfirmKeyboard returns the confirm/cancel keyboard for a rebuy operation.
func rebuyConfirmKeyboard(clubID, gameID, playerID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	pid := strconv.FormatInt(playerID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить Rebuy", fmt.Sprintf("%s:%s:%s:%s", cbGameRebuyConfirm, cid, gid, pid)),
			tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameRebuy, cid, gid)),
		),
	)
}

// rebuyManagePlayerSelectKeyboard builds an inline keyboard for selecting a player
// to manage/fix rebuy operations.
func rebuyManagePlayerSelectKeyboard(clubID, gameID int64, participants []*domain.GameParticipantWithPlayer) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(participants)+1)

	for _, p := range participants {
		label := p.Player.FirstName
		if p.Player.LastName != "" {
			label += " " + p.Player.LastName
		}
		if p.Player.Nickname != "" && p.Player.Nickname != p.Player.FirstName {
			label += " (@" + p.Player.Nickname + ")"
		}
		label += fmt.Sprintf(" [Rebuy: %d]", p.RebuyCount)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%s:%s:%d", cbGameRebuyFixOp, cid, gid, p.PlayerID)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameActiveBack, cid, gid)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// rebuyFixOpKeyboard builds an inline keyboard for selecting a rebuy event to fix.
func rebuyFixOpKeyboard(clubID, gameID, playerID int64, events []*domain.Event) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	pid := strconv.FormatInt(playerID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(events)+1)

	for _, e := range events {
		label := fmt.Sprintf("Rebuy #%d — %s", e.ID, formatEventValue(e.NewValue))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%s:%s:%s:%d", cbGameRebuyFixConfirm, cid, gid, pid, e.ID)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameRebuyManage, cid, gid)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// rebuyFixConfirmKeyboard returns the keyboard for confirming a rebuy fix.
func rebuyFixConfirmKeyboard(clubID, gameID, playerID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameRebuyManage, cid, gid)),
		),
	)
}

// addPlayerSelectKeyboard builds an inline keyboard for selecting a club member to add to the game.
func addPlayerSelectKeyboard(clubID, gameID int64, members []*domain.ClubMemberWithPlayer) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(members)+1)

	for _, m := range members {
		if m.Status != "active" {
			continue
		}
		label := m.Player.FirstName
		if m.Player.LastName != "" {
			label += " " + m.Player.LastName
		}
		if m.Player.Nickname != "" && m.Player.Nickname != m.Player.FirstName {
			label += " (@" + m.Player.Nickname + ")"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%s:%s:%d", cbGameAddPlayerConfirm, cid, gid, m.PlayerID)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameActiveBack, cid, gid)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// timerControlKeyboard returns the pause/resume timer keyboard.
func timerControlKeyboard(clubID, gameID int64, isPaused bool) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	var btnText, cb string
	if isPaused {
		btnText = "Продолжить"
		cb = cbGameResumeTimer
	} else {
		btnText = "Пауза"
		cb = cbGamePauseTimer
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, fmt.Sprintf("%s:%s:%s", cb, cid, gid)),
		),
	)
}

// gameExtendKeyboard returns the extend game keyboard.
func gameExtendKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Продлить игру", fmt.Sprintf("%s:%s:%s", cbGameExtend, cid, gid)),
		),
	)
}

// gameExtendSelectKeyboard returns the extension duration selection keyboard.
func gameExtendSelectKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("30 мин", fmt.Sprintf("%s:%s:%s:30", cbGameExtendSelect, cid, gid)),
			tgbotapi.NewInlineKeyboardButtonData("60 мин", fmt.Sprintf("%s:%s:%s:60", cbGameExtendSelect, cid, gid)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("120 мин", fmt.Sprintf("%s:%s:%s:120", cbGameExtendSelect, cid, gid)),
		),
	)
}

// playerStackKeyboard returns the keyboard for entering current stack.
func playerStackKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameActiveBack, cid, gid)),
		),
	)
}

// playerStatsKeyboard returns the keyboard for the player statistics view.
func playerStatsKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameActiveBack, cid, gid)),
		),
	)
}

// gameStatsKeyboard returns the keyboard for the game statistics view.
func gameStatsKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameActiveBack, cid, gid)),
		),
	)
}

// gameBankKeyboard returns the keyboard for the current bank view.
func gameBankKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameMonitor, cid, gid)),
		),
	)
}

// gameExpensesKeyboard returns the keyboard for the player expenses view.
func gameExpensesKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameMonitor, cid, gid)),
		),
	)
}

// formatEventValue formats a float64 pointer for display.
func formatEventValue(v *float64) string {
	if v == nil {
		return "—"
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

// formatTimerDisplay formats the remaining time for display.
func formatTimerDisplay(game *domain.Game) string {
	if game.Duration == nil {
		return ""
	}

	elapsed := time.Since(game.StartTime)
	pausedDuration := time.Duration(0)
	if game.TimerPausedDuration != nil {
		pausedDuration = *game.TimerPausedDuration
	}
	if game.TimerPausedAt != nil {
		pausedDuration += time.Since(*game.TimerPausedAt)
	}

	remaining := *game.Duration - elapsed + pausedDuration
	if remaining < 0 {
		remaining = 0
	}

	hours := int(remaining.Hours())
	minutes := int(remaining.Minutes()) % 60
	seconds := int(remaining.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// --- Phase 05: Game end keyboards ---

// gameEndKeyboard returns the inline keyboard for the game end screen.
// Shows buttons for entering chips_end, checking the bank, and confirming
// game completion. The "Завершить игру" button is only shown when all
// participants have chips_end set.
func gameEndKeyboard(clubID, gameID int64, allChipsEntered bool) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 5)

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Ввести chips_end", fmt.Sprintf("%s:%s:%s", cbGameEndPlayer, cid, gid)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Проверить банк", fmt.Sprintf("%s:%s:%s", cbGameEndCheckBank, cid, gid)),
	))

	if allChipsEntered {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Завершить игру", fmt.Sprintf("%s:%s:%s", cbGameEndConfirm, cid, gid)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameActiveBack, cid, gid)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// gameEndPlayerSelectKeyboard builds an inline keyboard for selecting a player
// to enter chips_end for the game end flow.
func gameEndPlayerSelectKeyboard(clubID, gameID int64, participants []*domain.GameParticipantWithPlayer) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(participants)+1)

	for _, p := range participants {
		label := p.Player.FirstName
		if p.Player.LastName != "" {
			label += " " + p.Player.LastName
		}
		if p.Player.Nickname != "" && p.Player.Nickname != p.Player.FirstName {
			label += " (@" + p.Player.Nickname + ")"
		}
		chipsStr := "—"
		if p.ChipsEnd != nil {
			chipsStr = formatFloat(*p.ChipsEnd)
		}
		label += fmt.Sprintf(" [%s]", chipsStr)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("%s:%s:%s:%d", cbGameEndPlayer, cid, gid, p.PlayerID)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameEnd, cid, gid)),
	))

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// gameEndConfirmKeyboard returns the confirm/cancel keyboard for game completion.
func gameEndConfirmKeyboard(clubID, gameID int64) tgbotapi.InlineKeyboardMarkup {
	cid := strconv.FormatInt(clubID, 10)
	gid := strconv.FormatInt(gameID, 10)
	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("Да", fmt.Sprintf("%s:%s:%s", cbGameEndFinish, cid, gid)),
				tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("%s:%s:%s", cbGameEnd, cid, gid)),
			},
		},
	}
}
