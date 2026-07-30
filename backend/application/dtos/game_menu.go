package dtos

import "time"

// InlineButton represents a Telegram inline keyboard button.
type InlineButton struct {
	Text    string `json:"text"`
	Action  string `json:"action"`
	Data    string `json:"data,omitempty"`
	Enabled bool   `json:"enabled"`
}

// GameMenuResponse is returned by ProcessGameMenu.
type GameMenuResponse struct {
	GameID       int64          `json:"game_id"`
	MenuType     string         `json:"menu_type"`
	Message      string         `json:"message"`
	Buttons      []InlineButton `json:"buttons"`
	CurrentState *GameState     `json:"current_state,omitempty"`
}

// GameState holds runtime state of a game.
type GameState struct {
	TimeLeft     time.Duration `json:"time_left"`
	TotalBank    float64       `json:"total_bank"`
	Participants []Participant `json:"participants"`
	GameType     string        `json:"game_type"`
	Status       string        `json:"status"`
}

// Participant represents a player in the game.
type Participant struct {
	PlayerID   int64   `json:"player_id"`
	PlayerName string  `json:"player_name"`
	BuyInCount int     `json:"buy_in_count"`
	RebuyCount int     `json:"rebuy_count"`
	TotalSpent float64 `json:"total_spent"`
	ChipsEnd   float64 `json:"chips_end"`
}

// GameStatsResponse is used for /game_stats command.
type GameStatsResponse struct {
	GameID       int64              `json:"game_id"`
	TimeLeft     time.Duration      `json:"time_left"`
	TotalBank    float64            `json:"total_bank"`
	Participants []ParticipantStats `json:"participants"`
	GameType     string             `json:"game_type"`
	Duration     time.Duration      `json:"duration"`
}

// ParticipantStats holds statistics for a participant.
type ParticipantStats struct {
	PlayerID   int     `json:"player_id"`
	PlayerName string  `json:"player_name"`
	BuyInCount int     `json:"buy_in_count"`
	RebuyCount int     `json:"rebuy_count"`
	TotalSpent float64 `json:"total_spent"`
	ChipsEnd   float64 `json:"chips_end"`
}
