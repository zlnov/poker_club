package persistence

import (
	"time"
)

// Club model for GORM
type Club struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for Club
func (Club) TableName() string {
	return "clubs"
}

// Player model for GORM
type Player struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	FirstName   string    `json:"first_name" gorm:"not null"`
	LastName    string    `json:"last_name" gorm:"not null"`
	Nickname    string    `json:"nickname"`
	PhoneNumber string    `json:"phone_number" gorm:"unique;not null"`
	Password    string    `json:"-" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for Player
func (Player) TableName() string {
	return "players"
}

// ClubMember model for GORM
type ClubMember struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ClubID    int64     `json:"club_id" gorm:"not null;index"`
	PlayerID  int64     `json:"player_id" gorm:"not null;index"`
	Role      string    `json:"role" gorm:"not null;check:role IN ('admin','member')"`
	Status    string    `json:"status" gorm:"not null;check:status IN ('pending','active','banned')"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Associations
	Club   Club   `json:"club" gorm:"foreignKey:ClubID;constraint:OnDelete:CASCADE;"`
	Player Player `json:"player" gorm:"foreignKey:PlayerID;constraint:OnDelete:CASCADE;"`
}

// TableName returns the table name for ClubMember
func (ClubMember) TableName() string {
	return "club_members"
}

// Game model for GORM
type Game struct {
	ID                 int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	ClubID             int64          `json:"club_id" gorm:"not null;index"`
	BankerID           int64          `json:"banker_id" gorm:"not null"`
	Type               string         `json:"type" gorm:"not null;check:type IN ('cash_time','cash_open','tournament')"`
	MoneyModel         string         `json:"money_model" gorm:"not null;default:'real'"`
	BuyInAmount        float64        `json:"buy_in_amount" gorm:"not null"`
	RebuyAllowed       bool           `json:"rebuy_allowed" gorm:"not null;default:false"`
	RebuyAmount        *float64       `json:"rebuy_amount"`
	MaxRebuysPerPlayer *int           `json:"max_rebuys_per_player"`
	Duration           *time.Duration `json:"duration"`
	StartTime          time.Time      `json:"start_time" gorm:"not null"`
	EndTime            *time.Time     `json:"end_time"`
	MinPlayers         int            `json:"min_players" gorm:"not null"`
	MaxPlayers         int            `json:"max_players" gorm:"not null"`
	RankingPrimary     string         `json:"ranking_primary" gorm:"not null"`
	RankingSecondary   *string        `json:"ranking_secondary"`
	CreatedAt          time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time      `json:"updated_at" gorm:"autoUpdateTime"`

	// Associations
	Club Club `json:"club" gorm:"foreignKey:ClubID;constraint:OnDelete:CASCADE;"`
}

// TableName returns the table name for Game
func (Game) TableName() string {
	return "games"
}

// GameParticipant model for GORM
type GameParticipant struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	GameID     int64     `json:"game_id" gorm:"not null;index;constraint:OnDelete:CASCADE;"`
	PlayerID   int64     `json:"player_id" gorm:"not null;index"`
	BuyInCount int       `json:"buy_in_count" gorm:"not null;default:0"`
	RebuyCount int       `json:"rebuy_count" gorm:"not null;default:0"`
	ChipsEnd   *float64  `json:"chips_end"`
	Place      *int      `json:"place"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Associations
	Game   Game   `json:"game" gorm:"foreignKey:GameID;"`
	Player Player `json:"player" gorm:"foreignKey:PlayerID;"`
}

// TableName returns the table name for GameParticipant
func (GameParticipant) TableName() string {
	return "game_participants"
}

// Event model for GORM
type Event struct {
	ID        int64                  `json:"id" gorm:"primaryKey;autoIncrement"`
	GameID    int64                  `json:"game_id" gorm:"not null;index;constraint:OnDelete:CASCADE;"`
	PlayerID  int64                  `json:"player_id" gorm:"not null;index"`
	Type      string                 `json:"type" gorm:"not null;check:type IN ('buy_in','rebuy','chips_set','correction')"`
	Amount    *float64               `json:"amount"`
	Metadata  map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time              `json:"created_at" gorm:"autoCreateTime"`
	CreatedBy int64                  `json:"created_by" gorm:"not null"`

	// Associations
	Game Game `json:"game" gorm:"foreignKey:GameID;"`
}

// TableName returns the table name for Event
func (Event) TableName() string {
	return "events"
}
