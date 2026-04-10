package usecases

import (
	"context"
	//"fmt"

	"poker-club-backend/application/dtos"
	"poker-club-backend/domain"
	"poker-club-backend/infrastructure/persistence"
	"poker-club-backend/infrastructure/services"
)

// AuthUseCase handles authentication use cases
type AuthUseCase struct {
	playerRepo domain.PlayerRepository
	jwtService *services.JWTService
}

// NewAuthUseCase creates a new AuthUseCase
func NewAuthUseCase(playerRepo domain.PlayerRepository, jwtService *services.JWTService) *AuthUseCase {
	return &AuthUseCase{
		playerRepo: playerRepo,
		jwtService: jwtService,
	}
}

// Login authenticates a user and returns tokens
func (uc *AuthUseCase) Login(ctx context.Context, req dtos.LoginRequest) (*dtos.LoginResponse, error) {
	// Find player by phone number
	player, err := uc.playerRepo.GetByPhone(ctx, req.PhoneNumber)
	if err != nil {
		return nil, domain.ErrPlayerNotFound
	}
	if player == nil {
		return nil, domain.ErrPlayerNotFound
	}

	// Check password using bcrypt
	if err := persistence.CheckPassword(player.Password, req.Password); err != nil {
		//fmt.Printf("Сверка паролей: Пароль пользователя: %s, Входящий пароль: %s", player.Password, req.Password)
		return nil, domain.ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := uc.jwtService.GenerateAccessToken(player.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.jwtService.GenerateRefreshToken(player.ID)
	if err != nil {
		return nil, err
	}

	return &dtos.LoginResponse{
		User: dtos.UserDTO{
			ID:        player.ID,
			FirstName: player.FirstName,
			LastName:  player.LastName,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Refresh generates a new access token using a refresh token
func (uc *AuthUseCase) Refresh(ctx context.Context, req dtos.RefreshRequest) (*dtos.RefreshResponse, error) {
	// Validate refresh token
	userID, err := uc.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new access token
	accessToken, err := uc.jwtService.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	refreshToken, err := uc.jwtService.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &dtos.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Logout is a no-op for MVP (stateless JWT)
func (uc *AuthUseCase) Logout(ctx context.Context) error {
	// MVP: no action needed, client discards tokens
	return nil
}
