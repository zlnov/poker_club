package dtos

import (
	"time"

	"poker-club-backend/domain"
)

// LoginResponse represents response after successful login
type LoginResponse struct {
	User         UserDTO `json:"user"`
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
}

// RefreshResponse represents response after token refresh
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserDTO represents a user in auth responses
type UserDTO struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// ClubResponse represents a club in API responses
type ClubResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GameResponse represents a game in API responses
type GameResponse struct {
	ID                 int64          `json:"id"`
	ClubID             int64          `json:"club_id"`
	BankerID           int64          `json:"banker_id"`
	Type               string         `json:"type"`
	MoneyModel         string         `json:"money_model"`
	BuyInAmount        float64        `json:"buy_in_amount"`
	RebuyAllowed       bool           `json:"rebuy_allowed"`
	RebuyAmount        *float64       `json:"rebuy_amount"`
	MaxRebuysPerPlayer *int           `json:"max_rebuys_per_player"`
	Duration           *time.Duration `json:"duration"`
	StartTime          time.Time      `json:"start_time"`
	EndTime            *time.Time     `json:"end_time"`
	MinPlayers         int            `json:"min_players"`
	MaxPlayers         int            `json:"max_players"`
	RankingPrimary     string         `json:"ranking_primary"`
	RankingSecondary   *string        `json:"ranking_secondary"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

// EventResponse represents an event in API responses
type EventResponse struct {
	ID        int64                  `json:"id"`
	GameID    int64                  `json:"game_id"`
	PlayerID  int64                  `json:"player_id"`
	Type      domain.EventType       `json:"type"`
	Amount    *float64               `json:"amount"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	CreatedBy int64                  `json:"created_by"`
}

// ToClubResponse converts domain.Club to ClubResponse
func ToClubResponse(club *domain.Club) *ClubResponse {
	return &ClubResponse{
		ID:        club.ID,
		Name:      club.Name,
		CreatedAt: club.CreatedAt,
		UpdatedAt: club.UpdatedAt,
	}
}

// ToGameResponse converts domain.Game to GameResponse
func ToGameResponse(game *domain.Game) *GameResponse {
	return &GameResponse{
		ID:                 game.ID,
		ClubID:             game.ClubID,
		BankerID:           game.BankerID,
		Type:               game.Type,
		MoneyModel:         game.MoneyModel,
		BuyInAmount:        game.BuyInAmount,
		RebuyAllowed:       game.RebuyAllowed,
		RebuyAmount:        game.RebuyAmount,
		MaxRebuysPerPlayer: game.MaxRebuysPerPlayer,
		Duration:           game.Duration,
		StartTime:          game.StartTime,
		EndTime:            game.EndTime,
		MinPlayers:         game.MinPlayers,
		MaxPlayers:         game.MaxPlayers,
		RankingPrimary:     game.RankingPrimary,
		RankingSecondary:   game.RankingSecondary,
		CreatedAt:          game.CreatedAt,
		UpdatedAt:          game.UpdatedAt,
	}
}

// ToEventResponse converts domain.Event to EventResponse
func ToEventResponse(event *domain.Event) *EventResponse {
	return &EventResponse{
		ID:        event.ID,
		GameID:    event.GameID,
		PlayerID:  event.PlayerID,
		Type:      domain.EventType(event.Type),
		Amount:    event.Amount,
		Metadata:  event.Metadata,
		CreatedAt: event.CreatedAt,
		CreatedBy: event.CreatedBy,
	}
}
