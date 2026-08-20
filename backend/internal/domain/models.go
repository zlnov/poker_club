package domain

import "time"

// Club represents a poker club.
type Club struct {
	ID        int64
	TgChatID  *int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Player represents a user in the system.
type Player struct {
	ID          int64
	FirstName   string
	LastName    string
	Nickname    string
	PhoneNumber string
	Email       string
	Password    string
	TgUserID    *int64
	LastSeen    time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ClubMember represents a player's membership in a club.
type ClubMember struct {
	ID        int64
	ClubID    int64
	PlayerID  int64
	Role      string // owner, admin, member
	Status    string // pending, active, banned, left
	Accepted  bool   // true when the invited user has accepted the invitation (pending → awaiting confirmation)
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ClubMemberWithPlayer represents a club member with their associated player info.
type ClubMemberWithPlayer struct {
	ClubMember
	Player Player
}

// Game represents a poker game session.
type Game struct {
	ID               int64
	ClubID           int64
	BankerID         int64
	GameType         string // cash, tournament
	Currency         string
	MoneyModel       string
	ChipValue        float64
	BuyInAmount      float64
	RebuyAllowed     bool
	RebuyPrice       *float64
	MaxRebuys        *int
	Duration         *time.Duration
	StartTime        time.Time
	EndTime          *time.Time
	Status           string // planned, active, finished, cancelled
	MinPlayers       int
	MaxPlayers       int
	RankingPrimary   string
	RankingSecondary *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	// Timer state (for Cash games with duration)
	TimerPausedAt       *time.Time
	TimerPausedDuration *time.Duration
	TimerNotified       bool
}

// GameParticipant represents a player registered in a game.
type GameParticipant struct {
	ID           int64
	GameID       int64
	PlayerID     int64
	BuyInCount   int
	RebuyCount   int
	ChipsEnd     *float64
	PayoutAmount *float64
	Place        *int
	Status       string // invited, accepted, declined, confirmed
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// GameParticipantWithPlayer represents a game participant with their associated player info.
type GameParticipantWithPlayer struct {
	GameParticipant
	Player Player
}

// Event represents an entry in the event log.
type Event struct {
	ID        int64
	GameID    int64
	PlayerID  int64
	Type      string // buy_in, rebuy, chips_set, correction
	OldValue  *float64
	NewValue  *float64
	Metadata  map[string]interface{}
	CreatedAt time.Time
	CreatedBy int64
}

// PlayerStatistics represents cached aggregate statistics for a player in a club.
type PlayerStatistics struct {
	ID               int64
	PlayerID         int64
	ClubID           int64
	TotalGames       int
	TotalBuyInAmount float64
	TotalRebuyAmount float64
	TotalRebuysCount int
	TotalInvested    float64
	TotalChips       float64
	TotalProfit      float64
	BiggestWin       float64
	BiggestLoss      float64
	GamesWon         int
	Podiums          int
	ROI              float64
	ITM              float64
	UpdatedAt        time.Time
}
