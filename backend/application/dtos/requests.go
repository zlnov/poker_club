package dtos

import (
	"time"

	"poker-club-backend/domain"
)

// LoginRequest represents request to login
type LoginRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Password    string `json:"password" validate:"required"`
}

// RefreshRequest represents request to refresh token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// CreateClubRequest represents request to create a club
type CreateClubRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=100"`
	CreatorID int64  `json:"-"`
}

// CreateClubResponse represents response after creating a club
type CreateClubResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ApproveMemberRequest represents request to approve a member
type ApproveMemberRequest struct {
	ClubID     int64 `json:"club_id" validate:"required"`
	MemberID   int64 `json:"member_id" validate:"required"`
	ApproverID int64 `json:"-"`
}

// RejectMemberRequest represents request to reject a member
type RejectMemberRequest struct {
	ClubID     int64 `json:"club_id" validate:"required"`
	MemberID   int64 `json:"member_id" validate:"required"`
	RejecterID int64 `json:"-"`
}

// GetClubMembersResponse represents response with club members
type GetClubMembersResponse struct {
	Members []ClubMemberDTO `json:"members"`
}

// ClubMemberDTO represents a club member in responses
type ClubMemberDTO struct {
	ID       int64               `json:"id"`
	ClubID   int64               `json:"club_id"`
	PlayerID int64               `json:"player_id"`
	Role     domain.Role         `json:"role"`
	Status   domain.MemberStatus `json:"status"`
	Player   PlayerDTO           `json:"player"`
}

// PlayerDTO represents a player in responses
type PlayerDTO struct {
	ID          int64     `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Nickname    string    `json:"nickname"`
	PhoneNumber string    `json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateGameRequest represents request to create a game
type CreateGameRequest struct {
	ClubID             int64             `json:"club_id" validate:"required"`
	BankerID           int64             `json:"banker_id" validate:"required"`
	Type               domain.GameType   `json:"type" validate:"required"`
	MoneyModel         domain.MoneyModel `json:"money_model" validate:"required"`
	BuyInAmount        float64           `json:"buy_in_amount" validate:"required,gt=0"`
	RebuyAllowed       bool              `json:"rebuy_allowed"`
	RebuyAmount        *float64          `json:"rebuy_amount"`
	MaxRebuysPerPlayer *int              `json:"max_rebuys_per_player"`
	Duration           *time.Duration    `json:"duration"`
	StartTime          time.Time         `json:"start_time" validate:"required"`
	MinPlayers         int               `json:"min_players" validate:"required,gt=0"`
	MaxPlayers         int               `json:"max_players" validate:"required,gt=0"`
	RankingPrimary     string            `json:"ranking_primary" validate:"required"`
	RankingSecondary   *string           `json:"ranking_secondary"`
}

// CreateGameResponse represents response after creating a game
type CreateGameResponse struct {
	ID        int64     `json:"id"`
	ClubID    int64     `json:"club_id"`
	BankerID  int64     `json:"banker_id"`
	Type      string    `json:"type"`
	StartTime time.Time `json:"start_time"`
}

// BuyInRequest represents request for buy-in/rebuy
type BuyInRequest struct {
	GameID      int64 `json:"game_id" validate:"required"`
	PlayerID    int64 `json:"player_id" validate:"required"`
	PerformedBy int64 `json:"-"`
}

// BuyInResponse represents response after buy-in
type BuyInResponse struct {
	GameID     int64   `json:"game_id"`
	PlayerID   int64   `json:"player_id"`
	BuyInCount int     `json:"buy_in_count"`
	RebuyCount int     `json:"rebuy_count"`
	ChipsEnd   float64 `json:"chips_end"`
}

// SetChipsRequest represents request to set final chips
type SetChipsRequest struct {
	GameID      int64   `json:"game_id" validate:"required"`
	PlayerID    int64   `json:"player_id" validate:"required"`
	Chips       float64 `json:"chips" validate:"required,gte=0"`
	PerformedBy int64   `json:"-"`
}

// FinishGameRequest represents request to finish a game
type FinishGameRequest struct {
	GameID      int64 `json:"game_id" validate:"required"`
	PerformedBy int64 `json:"-"`
}

// FinishGameResponse represents response after finishing a game
type FinishGameResponse struct {
	GameID        int64     `json:"game_id"`
	EndTime       time.Time `json:"end_time"`
	TotalInvested float64   `json:"total_invested"`
	TotalChips    float64   `json:"total_chips"`
}

// GetGameParticipantsResponse represents response with game participants
type GetGameParticipantsResponse struct {
	GameID       int64                `json:"game_id"`
	Participants []GameParticipantDTO `json:"participants"`
}

// GameParticipantDTO represents a game participant in responses
type GameParticipantDTO struct {
	ID         int64   `json:"id"`
	PlayerID   int64   `json:"player_id"`
	PlayerName string  `json:"player_name"`
	BuyInCount int     `json:"buy_in_count"`
	RebuyCount int     `json:"rebuy_count"`
	ChipsEnd   float64 `json:"chips_end"`
	Place      *int    `json:"place"`
}

// GetPlayerStatsResponse represents player statistics
type GetPlayerStatsResponse struct {
	PlayerID      int64   `json:"player_id"`
	TotalGames    int     `json:"total_games"`
	TotalBuyIn    float64 `json:"total_buy_in"`
	TotalRebuy    float64 `json:"total_rebuy"`
	TotalInvested float64 `json:"total_invested"`
	TotalChips    float64 `json:"total_chips"`
	Profit        float64 `json:"profit"`
	ROI           float64 `json:"roi"`
	ITM           float64 `json:"itm"` // In The Money percentage
}

// ListGamesRequest represents request to list games
type ListGamesRequest struct {
	ClubID int64  `json:"club_id" validate:"required"`
	Status string `json:"status" validate:"omitempty,oneof=active finished"`
	Limit  int    `json:"limit" validate:"required,min=1,max=100"`
	Offset int    `json:"offset" validate:"min=0"`
}

// ListGamesResponse represents response with list of games
type ListGamesResponse struct {
	Games  []GameListItemDTO `json:"games"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

// GameListItemDTO represents a game in list view (minimal data)
type GameListItemDTO struct {
	ID           int64      `json:"id"`
	ClubID       int64      `json:"club_id"`
	Type         string     `json:"type"`
	MoneyModel   string     `json:"money_model"`
	BuyInAmount  float64    `json:"buy_in_amount"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Participants int        `json:"participants_count"`
}

// GetGameDetailsResponse represents response with game details and participants
type GetGameDetailsResponse struct {
	Game         *GameResponse              `json:"game"`
	Participants []GameParticipantDetailDTO `json:"participants"`
}

// GameParticipantDetailDTO represents a participant with player details
type GameParticipantDetailDTO struct {
	ID         int64   `json:"id"`
	PlayerID   int64   `json:"player_id"`
	PlayerName string  `json:"player_name"`
	BuyInCount int     `json:"buy_in_count"`
	RebuyCount int     `json:"rebuy_count"`
	ChipsEnd   float64 `json:"chips_end"`
	Place      *int    `json:"place"`
}

// GetLeaderboardRequest represents request for leaderboard
type GetLeaderboardRequest struct {
	ClubID int64  `json:"club_id" validate:"required"`
	Metric string `json:"metric" validate:"required,oneof=profit roi winrate"`
	Period string `json:"period" validate:"required,oneof=all"`
}

// LeaderboardEntryDTO represents a single leaderboard entry
type LeaderboardEntryDTO struct {
	PlayerID   int64   `json:"player_id"`
	PlayerName string  `json:"player_name"`
	Metric     float64 `json:"metric_value"`
	GamesCount int     `json:"games_count"`
}

// GetLeaderboardResponse represents leaderboard response
type GetLeaderboardResponse struct {
	Metric   string                `json:"metric"`
	Period   string                `json:"period"`
	Entries  []LeaderboardEntryDTO `json:"entries"`
	ClubID   int64                 `json:"club_id"`
	ClubName string                `json:"club_name"`
}
