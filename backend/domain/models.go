package domain

import "time"

// Club represents a poker club (chat)
type Club struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Player represents a user
type Player struct {
	ID          int64     `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Nickname    string    `json:"nickname"`
	PhoneNumber string    `json:"phone_number"`
	Password    string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ClubMember links a player to a club with role and status
type ClubMember struct {
	ID        int64     `json:"id"`
	ClubID    int64     `json:"club_id"`
	PlayerID  int64     `json:"player_id"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Game represents a poker game
type Game struct {
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

// GameParticipant represents a player in a game
type GameParticipant struct {
	ID         int64     `json:"id"`
	GameID     int64     `json:"game_id"`
	PlayerID   int64     `json:"player_id"`
	BuyInCount int       `json:"buy_in_count"`
	RebuyCount int       `json:"rebuy_count"`
	ChipsEnd   *float64  `json:"chips_end"`
	Place      *int      `json:"place"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Event represents an audit event
type Event struct {
	ID        int64                  `json:"id"`
	GameID    int64                  `json:"game_id"`
	PlayerID  int64                  `json:"player_id"`
	Type      string                 `json:"type"`
	Amount    *float64               `json:"amount"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	CreatedBy int64                  `json:"created_by"`
}

// ChatClub maps a Telegram chat ID to a club ID
type ChatClub struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chat_id"`
	ClubID    int64     `json:"club_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
