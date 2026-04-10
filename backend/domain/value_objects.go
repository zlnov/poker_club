package domain

import (
	"errors"
	"fmt"
)

// Role represents a club member role
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

// MemberStatus represents a club member status
type MemberStatus string

const (
	StatusPending MemberStatus = "pending"
	StatusActive  MemberStatus = "active"
	StatusBanned  MemberStatus = "banned"
)

// GameType represents the type of poker game
type GameType string

const (
	GameTypeCashTime   GameType = "cash_time"
	GameTypeCashOpen   GameType = "cash_open"
	GameTypeTournament GameType = "tournament"
)

// MoneyModel represents the money model
type MoneyModel string

const (
	MoneyModelReal MoneyModel = "real"
	MoneyModelChip MoneyModel = "chip"
)

// EventType represents the type of event
type EventType string

const (
	EventBuyIn      EventType = "buy_in"
	EventRebuy      EventType = "rebuy"
	EventChipsSet   EventType = "chips_set"
	EventCorrection EventType = "correction"
)

// Validate checks if the role is valid
func (r Role) Validate() error {
	switch r {
	case RoleAdmin, RoleMember:
		return nil
	default:
		return fmt.Errorf("invalid role: %s", r)
	}
}

// Validate checks if the status is valid
func (s MemberStatus) Validate() error {
	switch s {
	case StatusPending, StatusActive, StatusBanned:
		return nil
	default:
		return fmt.Errorf("invalid status: %s", s)
	}
}

// Validate checks if the game type is valid
func (gt GameType) Validate() error {
	switch gt {
	case GameTypeCashTime, GameTypeCashOpen, GameTypeTournament:
		return nil
	default:
		return fmt.Errorf("invalid game type: %s", gt)
	}
}

// Validate checks if the event type is valid
func (et EventType) Validate() error {
	switch et {
	case EventBuyIn, EventRebuy, EventChipsSet, EventCorrection:
		return nil
	default:
		return fmt.Errorf("invalid event type: %s", et)
	}
}

// Validate checks if the money model is valid
func (mm MoneyModel) Validate() error {
	switch mm {
	case MoneyModelReal, MoneyModelChip:
		return nil
	default:
		return fmt.Errorf("invalid money model: %s", mm)
	}
}

// Domain errors
var (
	ErrClubNotFound         = errors.New("club not found")
	ErrPlayerNotFound       = errors.New("player not found")
	ErrMemberNotFound       = errors.New("club member not found")
	ErrGameNotFound         = errors.New("game not found")
	ErrParticipantNotFound  = errors.New("game participant not found")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrInvalidRole          = errors.New("invalid role")
	ErrInvalidStatus        = errors.New("invalid status")
	ErrInvalidGameType      = errors.New("invalid game type")
	ErrInvalidEventType     = errors.New("invalid event type")
	ErrGameAlreadyEnded     = errors.New("game has already ended")
	ErrGameNotStarted       = errors.New("game has not started")
	ErrRebuyNotAllowed      = errors.New("rebuy is not allowed for this game")
	ErrMaxRebuysExceeded    = errors.New("maximum number of rebuys exceeded")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidChipsAmount   = errors.New("invalid chips amount")
	ErrChipsMismatch        = errors.New("sum of chips_end does not equal total invested")
	ErrDuplicateParticipant = errors.New("player is already a participant in this game")
)
