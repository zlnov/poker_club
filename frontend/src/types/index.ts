/**
 * Domain types for the Poker Club Frontend.
 *
 * These types represent the data models used by the Frontend application.
 * They are derived from the Backend domain models (02_DB.md) and the
 * API specification (06_API.md).
 *
 * Frontend does not contain business logic — it only displays state
 * provided by the Backend and shown via these types.
 *
 * See 04_FE_SPEC.md section 17 (Business Logic) and
 * section 18 (Game State).
 */

import type { ApiErrorCode } from '../api'

/** Player identity and profile information. */
export interface PlayerSummary {
  /** Telegram user ID. */
  tgUserID: number
  /** First name. */
  firstName: string
  /** Last name. */
  lastName: string
  /** Nickname/username. */
  nickname: string
  /** Profile photo URL (optional). */
  photoUrl?: string
}

/** Club information. */
export interface ClubSummary {
  /** Club ID. */
  id: number
  /** Club name. */
  name: string
  /** Telegram chat ID (if bound). */
  tgChatId?: number
  /** Number of members. */
  memberCount: number
  /** Whether current user is the owner. */
  isOwner: boolean
  /** Whether current user is an admin. */
  isAdmin: boolean
}

/** Game type. */
export type GameType = 'cash_time' | 'cash_open' | 'tournament'

/** Game status. */
export type GameStatus = 'planned' | 'active' | 'finished' | 'cancelled'

/** Game configuration summary. */
export interface GameSummary {
  /** Game ID. */
  id: number
  /** Club ID the game belongs to. */
  clubId: number
  /** Game type. */
  type: GameType
  /** Status. */
  status: GameStatus
  /** Banker player ID. */
  bankerId: number
  /** Banker display name. */
  bankerName: string
  /** Start time. */
  startTime: string
  /** End time (if finished). */
  endTime?: string
  /** Currency. */
  currency: string
  /** Chip value. */
  chipValue: number
  /** Buy-in amount. */
  buyInAmount: number
  /** Rebuy allowed. */
  rebuyAllowed: boolean
  /** Rebuy price (if allowed). */
  rebuyPrice?: number
  /** Max rebuys (if allowed). */
  maxRebuys?: number
  /** Duration in seconds (for cash_time games). */
  durationSeconds?: number
  /** Min players. */
  minPlayers: number
  /** Max players. */
  maxPlayers: number
  /** Current player count. */
  currentPlayers: number
}

/** Full game details from the Backend. */
export interface GameDetails {
  /** Game ID. */
  id: number
  /** Club ID the game belongs to. */
  clubId: number
  /** Banker club member ID. */
  bankerId: number
  /** Game type (cash or tournament). */
  gameType: string
  /** Currency code. */
  currency: string
  /** Money model (real, points, virtual, practice). */
  moneyModel: string
  /** Chip value. */
  chipValue: number
  /** Buy-in amount. */
  buyInAmount: number
  /** Rebuy allowed. */
  rebuyAllowed: boolean
  /** Rebuy price (if allowed). */
  rebuyPrice?: number
  /** Max rebuys (if allowed). */
  maxRebuys?: number
  /** Duration in seconds (for cash_time games). */
  durationSeconds?: number
  /** Start time. */
  startTime: string
  /** End time (if finished). */
  endTime?: string
  /** Status. */
  status: GameStatus
  /** Min players. */
  minPlayers: number
  /** Max players. */
  maxPlayers: number
  /** Primary ranking method. */
  rankingPrimary: string
  /** Secondary ranking method (optional). */
  rankingSecondary?: string
  /** Created at timestamp. */
  createdAt: string
  /** Updated at timestamp. */
  updatedAt: string
}

/** Game configuration form values. */
export interface GameConfig {
  /** Game type. */
  gameType: GameType
  /** Currency code. */
  currency: string
  /** Money model. */
  moneyModel: string
  /** Chip value. */
  chipValue: number
  /** Buy-in amount. */
  buyInAmount: number
  /** Rebuy allowed. */
  rebuyAllowed: boolean
  /** Rebuy price (if allowed). */
  rebuyPrice?: number
  /** Max rebuys (if allowed). */
  maxRebuys?: number
  /** Duration in seconds (for cash_time games). */
  durationSeconds?: number
  /** Scheduled start time. */
  startTime?: string
  /** Min players. */
  minPlayers: number
  /** Max players. */
  maxPlayers: number
  /** Primary ranking method. */
  rankingPrimary: string
  /** Secondary ranking method (optional). */
  rankingSecondary?: string
  /** Banker player ID. */
  bankerId: number
}

/** Game participant information. */
export interface GameParticipantSummary {
  /** Player. */
  player: PlayerSummary
  /** Buy-in count. */
  buyInCount: number
  /** Rebuy count. */
  rebuyCount: number
  /** Ending chips (if game finished). */
  chipsEnd?: number
  /** Payout amount (if game finished). */
  payoutAmount?: number
  /** Place (if game finished). */
  place?: number
  /** Status (invited, accepted, declined, confirmed). */
  status: 'invited' | 'accepted' | 'declined' | 'confirmed'
}

/** Statistics for a player. */
export interface PlayerStatistics {
  /** Total games played. */
  totalGames: number
  /** Total buy-in amount. */
  totalBuyInAmount: number
  /** Total rebuy amount. */
  totalRebuyAmount: number
  /** Total invested (buy-in + rebuy). */
  totalInvested: number
  /** Total chips won/lost. */
  totalChips: number
  /** Total profit. */
  totalProfit: number
  /** Biggest win. */
  biggestWin: number
  /** Biggest loss. */
  biggestLoss: number
  /** Games won (profit > 0 for cash). */
  gamesWon: number
  /** Podiums (top 3 for tournament). */
  podiums: number
  /** ROI percentage. */
  roi: number
  /** ITM percentage (tournament only). */
  itm: number
  /** Win rate percentage. */
  winrate: number
  /** Average place. */
  avgPlace: number
}

/** Statistics for a club. */
export interface ClubStatistics {
  /** Total members. */
  totalMembers: number
  /** Total games. */
  totalGames: number
  /** Cash games count. */
  cashGames: number
  /** Tournament games count. */
  tournamentGames: number
  /** Total buy-in amount. */
  totalBuyInAmount: number
  /** Total rebuy amount. */
  totalRebuyAmount: number
  /** Total bank (sum of invested). */
  totalBank: number
  /** Average game duration. */
  averageGameDuration: string
}

/** Current authenticated user. */
export interface CurrentUser {
  /** User ID. */
  id: number
  /** Player summary. */
  player: PlayerSummary
  /** Club memberships. */
  clubMemberships: ClubMemberRole[]
  /** Current club ID (if any). */
  currentClubId?: number
  /** Current game ID (if any). */
  currentGameId?: number
}

/** Club member role. */
export interface ClubMemberRole {
  /** Club member ID (maps to club_members.id in Backend). */
  clubMemberId: number
  /** Player ID. */
  playerId: number
  /** Player first name. */
  firstName: string
  /** Player last name. */
  lastName: string
  /** Player nickname/username. */
  nickname: string
  /** Role (owner, admin, member). */
  role: 'owner' | 'admin' | 'member'
  /** Status (pending, active, banned, left). */
  status: 'pending' | 'active' | 'banned' | 'left'
  /** Can manage game participants. */
  canManageParticipants: boolean
  /** Can adjust game results. */
  canAdjustResults: boolean
  /** Can invite members. */
  canInvite: boolean
  /** Can remove members. */
  canRemove: boolean
}

/** Invitation information for a club. */
export interface InvitationInfo {
  /** Player ID of the invited user. */
  playerId: number
  /** Player first name. */
  firstName: string
  /** Player last name. */
  lastName: string
  /** Player nickname/username. */
  nickname: string
  /** Telegram user ID of the invited player. */
  tgUserId?: number
  /** Invitation status (always 'pending' for invitations). */
  status: 'pending'
  /** Whether the invitation has been accepted. */
  accepted: boolean
  /** When the invitation was created. */
  createdAt: string
  /** When the invitation was last updated. */
  updatedAt: string
}

/** API response wrapper. */
export interface ApiResponse<T> {
  /** Success flag. */
  success: boolean
  /** Data payload. */
  data: T
  /** Error code (if failed). */
  errorCode?: ApiErrorCode
  /** Human-readable error message. */
  errorMessage?: string
}

/** Pagination metadata. */
export interface PaginationMeta {
  /** Current page number. */
  page: number
  /** Items per page. */
  pageSize: number
  /** Total number of items. */
  totalCount: number
  /** Total number of pages. */
  totalPages: number
}

/** Paginated response. */
export interface PaginatedResponse<T> {
  /** Items list. */
  items: T[]
  /** Pagination metadata. */
  meta: PaginationMeta
}

/** Filter options for API requests. */
export interface ListFilters {
  /** Search query. */
  search?: string
  /** Sort field. */
  sortBy?: keyof any
  /** Sort direction. */
  sortDir?: 'asc' | 'desc'
  /** Page number. */
  page?: number
  /** Page size. */
  pageSize?: number
}

/** Sort configuration. */
export interface SortConfig {
  /** Field to sort by. */
  field: string
  /** Sort direction. */
  direction: 'asc' | 'desc'
}

/** Notification toast configuration. */
export interface ToastConfig {
  /** Toast message. */
  message: string
  /** Toast type (info, success, error, warning). */
  type: 'info' | 'success' | 'error' | 'warning'
  /** Duration in milliseconds. */
  duration?: number
}

/** Loading state. */
export type LoadingState = 'idle' | 'loading' | 'success' | 'error'
