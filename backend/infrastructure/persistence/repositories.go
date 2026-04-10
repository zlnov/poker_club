package persistence

import (
	"context"
	"errors"
	"fmt"

	"poker-club-backend/domain"

	"gorm.io/gorm"
)

// ClubRepository implementation
type ClubRepository struct {
	db *gorm.DB
}

func NewClubRepository(db *gorm.DB) domain.ClubRepository {
	return &ClubRepository{db: db}
}

func (r *ClubRepository) Create(ctx context.Context, club *domain.Club) error {
	model := &Club{
		Name: club.Name,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	club.ID = model.ID
	club.CreatedAt = model.CreatedAt
	club.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *ClubRepository) GetByID(ctx context.Context, id int64) (*domain.Club, error) {
	var model Club
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.Club{
		ID:        model.ID,
		Name:      model.Name,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *ClubRepository) GetAll(ctx context.Context) ([]*domain.Club, error) {
	var models []Club
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	clubs := make([]*domain.Club, 0, len(models))
	for _, m := range models {
		clubs = append(clubs, &domain.Club{
			ID:        m.ID,
			Name:      m.Name,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}
	return clubs, nil
}

func (r *ClubRepository) Update(ctx context.Context, club *domain.Club) error {
	model := &Club{
		ID:        club.ID,
		Name:      club.Name,
		UpdatedAt: club.UpdatedAt,
	}
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *ClubRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&Club{}, id).Error
}

// PlayerRepository implementation
type PlayerRepository struct {
	db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) domain.PlayerRepository {
	return &PlayerRepository{db: db}
}

func (r *PlayerRepository) Create(ctx context.Context, player *domain.Player) error {
	model := &Player{
		FirstName:   player.FirstName,
		LastName:    player.LastName,
		Nickname:    player.Nickname,
		PhoneNumber: player.PhoneNumber,
		Password:    player.Password,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	player.ID = model.ID
	player.CreatedAt = model.CreatedAt
	player.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *PlayerRepository) GetByID(ctx context.Context, id int64) (*domain.Player, error) {
	var model Player
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.Player{
		ID:          model.ID,
		FirstName:   model.FirstName,
		LastName:    model.LastName,
		Nickname:    model.Nickname,
		PhoneNumber: model.PhoneNumber,
		Password:    model.Password,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

func (r *PlayerRepository) GetByPhone(ctx context.Context, phoneNumber string) (*domain.Player, error) {
	var model Player

	fmt.Printf("2:На проверку пришел номер: %#v\n", phoneNumber)

	if err := r.db.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			fmt.Println("Проверка попадания в ошибку")

			return nil, nil
		}
		fmt.Printf("Проверка попадая в ошибку: %v\n", err)
		return nil, err
	}

	fmt.Printf("3:Пользователь полученный из Базы Данных %+v\n", model)

	return &domain.Player{
		ID:          model.ID,
		FirstName:   model.FirstName,
		LastName:    model.LastName,
		Nickname:    model.Nickname,
		PhoneNumber: model.PhoneNumber,
		Password:    model.Password,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

func (r *PlayerRepository) GetAll(ctx context.Context) ([]*domain.Player, error) {
	var models []Player
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	players := make([]*domain.Player, 0, len(models))
	for _, m := range models {
		players = append(players, &domain.Player{
			ID:          m.ID,
			FirstName:   m.FirstName,
			LastName:    m.LastName,
			Nickname:    m.Nickname,
			PhoneNumber: m.PhoneNumber,
			Password:    m.Password,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
		})
	}
	return players, nil
}

func (r *PlayerRepository) Update(ctx context.Context, player *domain.Player) error {
	model := &Player{
		ID:          player.ID,
		FirstName:   player.FirstName,
		LastName:    player.LastName,
		Nickname:    player.Nickname,
		PhoneNumber: player.PhoneNumber,
		Password:    player.Password,
		UpdatedAt:   player.UpdatedAt,
	}
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *PlayerRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&Player{}, id).Error
}

// ClubMemberRepository implementation
type ClubMemberRepository struct {
	db *gorm.DB
}

func NewClubMemberRepository(db *gorm.DB) domain.ClubMemberRepository {
	return &ClubMemberRepository{db: db}
}

func (r *ClubMemberRepository) Create(ctx context.Context, member *domain.ClubMember) error {
	model := &ClubMember{
		ClubID:   member.ClubID,
		PlayerID: member.PlayerID,
		Role:     member.Role,
		Status:   member.Status,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	member.ID = model.ID
	member.CreatedAt = model.CreatedAt
	member.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *ClubMemberRepository) GetByID(ctx context.Context, id int64) (*domain.ClubMember, error) {
	var model ClubMember
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.ClubMember{
		ID:        model.ID,
		ClubID:    model.ClubID,
		PlayerID:  model.PlayerID,
		Role:      model.Role,
		Status:    model.Status,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *ClubMemberRepository) GetByClubID(ctx context.Context, clubID int64) ([]*domain.ClubMember, error) {
	var models []ClubMember
	if err := r.db.WithContext(ctx).Where("club_id = ?", clubID).Find(&models).Error; err != nil {
		return nil, err
	}
	members := make([]*domain.ClubMember, 0, len(models))
	for _, m := range models {
		members = append(members, &domain.ClubMember{
			ID:        m.ID,
			ClubID:    m.ClubID,
			PlayerID:  m.PlayerID,
			Role:      m.Role,
			Status:    m.Status,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}
	return members, nil
}

func (r *ClubMemberRepository) GetByPlayerID(ctx context.Context, playerID int64) ([]*domain.ClubMember, error) {
	var models []ClubMember
	if err := r.db.WithContext(ctx).Where("player_id = ?", playerID).Find(&models).Error; err != nil {
		return nil, err
	}
	members := make([]*domain.ClubMember, 0, len(models))
	for _, m := range models {
		members = append(members, &domain.ClubMember{
			ID:        m.ID,
			ClubID:    m.ClubID,
			PlayerID:  m.PlayerID,
			Role:      m.Role,
			Status:    m.Status,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}
	return members, nil
}

func (r *ClubMemberRepository) GetByClubAndPlayer(ctx context.Context, clubID, playerID int64) (*domain.ClubMember, error) {
	var model ClubMember
	if err := r.db.WithContext(ctx).Where("club_id = ? AND player_id = ?", clubID, playerID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.ClubMember{
		ID:        model.ID,
		ClubID:    model.ClubID,
		PlayerID:  model.PlayerID,
		Role:      model.Role,
		Status:    model.Status,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *ClubMemberRepository) Update(ctx context.Context, member *domain.ClubMember) error {
	model := &ClubMember{
		ID:        member.ID,
		ClubID:    member.ClubID,
		PlayerID:  member.PlayerID,
		Role:      member.Role,
		Status:    member.Status,
		UpdatedAt: member.UpdatedAt,
	}
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *ClubMemberRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&ClubMember{}, id).Error
}

// GameRepository implementation
type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) domain.GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) Create(ctx context.Context, game *domain.Game) error {
	model := &Game{
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
		MinPlayers:         game.MinPlayers,
		MaxPlayers:         game.MaxPlayers,
		RankingPrimary:     game.RankingPrimary,
		RankingSecondary:   game.RankingSecondary,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	game.ID = model.ID
	game.CreatedAt = model.CreatedAt
	game.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *GameRepository) GetByID(ctx context.Context, id int64) (*domain.Game, error) {
	var model Game
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.Game{
		ID:                 model.ID,
		ClubID:             model.ClubID,
		BankerID:           model.BankerID,
		Type:               model.Type,
		MoneyModel:         model.MoneyModel,
		BuyInAmount:        model.BuyInAmount,
		RebuyAllowed:       model.RebuyAllowed,
		RebuyAmount:        model.RebuyAmount,
		MaxRebuysPerPlayer: model.MaxRebuysPerPlayer,
		Duration:           model.Duration,
		StartTime:          model.StartTime,
		EndTime:            model.EndTime,
		MinPlayers:         model.MinPlayers,
		MaxPlayers:         model.MaxPlayers,
		RankingPrimary:     model.RankingPrimary,
		RankingSecondary:   model.RankingSecondary,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}, nil
}

func (r *GameRepository) GetByClubID(ctx context.Context, clubID int64) ([]*domain.Game, error) {
	var models []Game
	if err := r.db.WithContext(ctx).Where("club_id = ?", clubID).Find(&models).Error; err != nil {
		return nil, err
	}
	games := make([]*domain.Game, 0, len(models))
	for _, m := range models {
		games = append(games, &domain.Game{
			ID:                 m.ID,
			ClubID:             m.ClubID,
			BankerID:           m.BankerID,
			Type:               m.Type,
			MoneyModel:         m.MoneyModel,
			BuyInAmount:        m.BuyInAmount,
			RebuyAllowed:       m.RebuyAllowed,
			RebuyAmount:        m.RebuyAmount,
			MaxRebuysPerPlayer: m.MaxRebuysPerPlayer,
			Duration:           m.Duration,
			StartTime:          m.StartTime,
			EndTime:            m.EndTime,
			MinPlayers:         m.MinPlayers,
			MaxPlayers:         m.MaxPlayers,
			RankingPrimary:     m.RankingPrimary,
			RankingSecondary:   m.RankingSecondary,
			CreatedAt:          m.CreatedAt,
			UpdatedAt:          m.UpdatedAt,
		})
	}
	return games, nil
}

func (r *GameRepository) GetByBankerID(ctx context.Context, bankerID int64) ([]*domain.Game, error) {
	var models []Game
	if err := r.db.WithContext(ctx).Where("banker_id = ?", bankerID).Find(&models).Error; err != nil {
		return nil, err
	}
	games := make([]*domain.Game, 0, len(models))
	for _, m := range models {
		games = append(games, &domain.Game{
			ID:                 m.ID,
			ClubID:             m.ClubID,
			BankerID:           m.BankerID,
			Type:               m.Type,
			MoneyModel:         m.MoneyModel,
			BuyInAmount:        m.BuyInAmount,
			RebuyAllowed:       m.RebuyAllowed,
			RebuyAmount:        m.RebuyAmount,
			MaxRebuysPerPlayer: m.MaxRebuysPerPlayer,
			Duration:           m.Duration,
			StartTime:          m.StartTime,
			EndTime:            m.EndTime,
			MinPlayers:         m.MinPlayers,
			MaxPlayers:         m.MaxPlayers,
			RankingPrimary:     m.RankingPrimary,
			RankingSecondary:   m.RankingSecondary,
			CreatedAt:          m.CreatedAt,
			UpdatedAt:          m.UpdatedAt,
		})
	}
	return games, nil
}

func (r *GameRepository) Update(ctx context.Context, game *domain.Game) error {
	model := &Game{
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
		UpdatedAt:          game.UpdatedAt,
	}
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *GameRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&Game{}, id).Error
}

// GetByClubIDWithFilters returns games for a club with optional status filter and pagination
func (r *GameRepository) GetByClubIDWithFilters(ctx context.Context, clubID int64, status *string, limit, offset int) ([]*domain.Game, error) {
	query := r.db.WithContext(ctx).Where("club_id = ?", clubID)

	// Apply status filter
	if status != nil {
		if *status == "active" {
			query = query.Where("end_time IS NULL")
		} else if *status == "finished" {
			query = query.Where("end_time IS NOT NULL")
		}
	}

	// Apply pagination
	query = query.Limit(limit).Offset(offset)

	var models []Game
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	games := make([]*domain.Game, 0, len(models))
	for _, m := range models {
		games = append(games, &domain.Game{
			ID:                 m.ID,
			ClubID:             m.ClubID,
			BankerID:           m.BankerID,
			Type:               m.Type,
			MoneyModel:         m.MoneyModel,
			BuyInAmount:        m.BuyInAmount,
			RebuyAllowed:       m.RebuyAllowed,
			RebuyAmount:        m.RebuyAmount,
			MaxRebuysPerPlayer: m.MaxRebuysPerPlayer,
			Duration:           m.Duration,
			StartTime:          m.StartTime,
			EndTime:            m.EndTime,
			MinPlayers:         m.MinPlayers,
			MaxPlayers:         m.MaxPlayers,
			RankingPrimary:     m.RankingPrimary,
			RankingSecondary:   m.RankingSecondary,
			CreatedAt:          m.CreatedAt,
			UpdatedAt:          m.UpdatedAt,
		})
	}
	return games, nil
}

// CountByClubID returns count of games for a club with optional status filter
func (r *GameRepository) CountByClubID(ctx context.Context, clubID int64, status *string) (int64, error) {
	query := r.db.WithContext(ctx).Model(&Game{}).Where("club_id = ?", clubID)

	// Apply status filter
	if status != nil {
		if *status == "active" {
			query = query.Where("end_time IS NULL")
		} else if *status == "finished" {
			query = query.Where("end_time IS NOT NULL")
		}
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GameParticipantRepository implementation
type GameParticipantRepository struct {
	db *gorm.DB
}

func NewGameParticipantRepository(db *gorm.DB) domain.GameParticipantRepository {
	return &GameParticipantRepository{db: db}
}

func (r *GameParticipantRepository) Create(ctx context.Context, participant *domain.GameParticipant) error {
	model := &GameParticipant{
		GameID:     participant.GameID,
		PlayerID:   participant.PlayerID,
		BuyInCount: participant.BuyInCount,
		RebuyCount: participant.RebuyCount,
		ChipsEnd:   participant.ChipsEnd,
		Place:      participant.Place,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	participant.ID = model.ID
	participant.CreatedAt = model.CreatedAt
	participant.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *GameParticipantRepository) GetByID(ctx context.Context, id int64) (*domain.GameParticipant, error) {
	var model GameParticipant
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.GameParticipant{
		ID:         model.ID,
		GameID:     model.GameID,
		PlayerID:   model.PlayerID,
		BuyInCount: model.BuyInCount,
		RebuyCount: model.RebuyCount,
		ChipsEnd:   model.ChipsEnd,
		Place:      model.Place,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}, nil
}

func (r *GameParticipantRepository) GetByGameID(ctx context.Context, gameID int64) ([]*domain.GameParticipant, error) {
	var models []GameParticipant
	if err := r.db.WithContext(ctx).Where("game_id = ?", gameID).Find(&models).Error; err != nil {
		return nil, err
	}
	participants := make([]*domain.GameParticipant, 0, len(models))
	for _, m := range models {
		participants = append(participants, &domain.GameParticipant{
			ID:         m.ID,
			GameID:     m.GameID,
			PlayerID:   m.PlayerID,
			BuyInCount: m.BuyInCount,
			RebuyCount: m.RebuyCount,
			ChipsEnd:   m.ChipsEnd,
			Place:      m.Place,
			CreatedAt:  m.CreatedAt,
			UpdatedAt:  m.UpdatedAt,
		})
	}
	return participants, nil
}

func (r *GameParticipantRepository) GetByPlayerID(ctx context.Context, playerID int64) ([]*domain.GameParticipant, error) {
	var models []GameParticipant
	if err := r.db.WithContext(ctx).Where("player_id = ?", playerID).Find(&models).Error; err != nil {
		return nil, err
	}
	participants := make([]*domain.GameParticipant, 0, len(models))
	for _, m := range models {
		participants = append(participants, &domain.GameParticipant{
			ID:         m.ID,
			GameID:     m.GameID,
			PlayerID:   m.PlayerID,
			BuyInCount: m.BuyInCount,
			RebuyCount: m.RebuyCount,
			ChipsEnd:   m.ChipsEnd,
			Place:      m.Place,
			CreatedAt:  m.CreatedAt,
			UpdatedAt:  m.UpdatedAt,
		})
	}
	return participants, nil
}

func (r *GameParticipantRepository) GetByGameAndPlayer(ctx context.Context, gameID, playerID int64) (*domain.GameParticipant, error) {
	var model GameParticipant
	if err := r.db.WithContext(ctx).Where("game_id = ? AND player_id = ?", gameID, playerID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.GameParticipant{
		ID:         model.ID,
		GameID:     model.GameID,
		PlayerID:   model.PlayerID,
		BuyInCount: model.BuyInCount,
		RebuyCount: model.RebuyCount,
		ChipsEnd:   model.ChipsEnd,
		Place:      model.Place,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}, nil
}

func (r *GameParticipantRepository) Update(ctx context.Context, participant *domain.GameParticipant) error {
	model := &GameParticipant{
		ID:         participant.ID,
		GameID:     participant.GameID,
		PlayerID:   participant.PlayerID,
		BuyInCount: participant.BuyInCount,
		RebuyCount: participant.RebuyCount,
		ChipsEnd:   participant.ChipsEnd,
		Place:      participant.Place,
		UpdatedAt:  participant.UpdatedAt,
	}
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *GameParticipantRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&GameParticipant{}, id).Error
}

// LockForUpdate locks all participants of a game for update to prevent race conditions
func (r *GameParticipantRepository) LockForUpdate(ctx context.Context, gameID int64) ([]*domain.GameParticipant, error) {
	var models []GameParticipant
	// Use raw SQL with SELECT FOR UPDATE to lock rows
	if err := r.db.WithContext(ctx).Raw("SELECT * FROM game_participants WHERE game_id = ? FOR UPDATE", gameID).Scan(&models).Error; err != nil {
		return nil, err
	}

	participants := make([]*domain.GameParticipant, 0, len(models))
	for _, m := range models {
		participants = append(participants, &domain.GameParticipant{
			ID:         m.ID,
			GameID:     m.GameID,
			PlayerID:   m.PlayerID,
			BuyInCount: m.BuyInCount,
			RebuyCount: m.RebuyCount,
			ChipsEnd:   m.ChipsEnd,
			Place:      m.Place,
			CreatedAt:  m.CreatedAt,
			UpdatedAt:  m.UpdatedAt,
		})
	}
	return participants, nil
}

// EventRepository implementation
type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) domain.EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, event *domain.Event) error {
	model := &Event{
		GameID:    event.GameID,
		PlayerID:  event.PlayerID,
		Type:      event.Type,
		Amount:    event.Amount,
		Metadata:  event.Metadata,
		CreatedBy: event.CreatedBy,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	event.ID = model.ID
	event.CreatedAt = model.CreatedAt
	return nil
}

func (r *EventRepository) GetByID(ctx context.Context, id int64) (*domain.Event, error) {
	var model Event
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.Event{
		ID:        model.ID,
		GameID:    model.GameID,
		PlayerID:  model.PlayerID,
		Type:      model.Type,
		Amount:    model.Amount,
		Metadata:  model.Metadata,
		CreatedAt: model.CreatedAt,
		CreatedBy: model.CreatedBy,
	}, nil
}

func (r *EventRepository) GetByGameID(ctx context.Context, gameID int64) ([]*domain.Event, error) {
	var models []Event
	if err := r.db.WithContext(ctx).Where("game_id = ?", gameID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	events := make([]*domain.Event, 0, len(models))
	for _, m := range models {
		events = append(events, &domain.Event{
			ID:        m.ID,
			GameID:    m.GameID,
			PlayerID:  m.PlayerID,
			Type:      m.Type,
			Amount:    m.Amount,
			Metadata:  m.Metadata,
			CreatedAt: m.CreatedAt,
			CreatedBy: m.CreatedBy,
		})
	}
	return events, nil
}

func (r *EventRepository) GetByPlayerID(ctx context.Context, playerID int64) ([]*domain.Event, error) {
	var models []Event
	if err := r.db.WithContext(ctx).Where("player_id = ?", playerID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	events := make([]*domain.Event, 0, len(models))
	for _, m := range models {
		events = append(events, &domain.Event{
			ID:        m.ID,
			GameID:    m.GameID,
			PlayerID:  m.PlayerID,
			Type:      m.Type,
			Amount:    m.Amount,
			Metadata:  m.Metadata,
			CreatedAt: m.CreatedAt,
			CreatedBy: m.CreatedBy,
		})
	}
	return events, nil
}

func (r *EventRepository) GetByGameAndPlayer(ctx context.Context, gameID, playerID int64) ([]*domain.Event, error) {
	var models []Event
	if err := r.db.WithContext(ctx).Where("game_id = ? AND player_id = ?", gameID, playerID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	events := make([]*domain.Event, 0, len(models))
	for _, m := range models {
		events = append(events, &domain.Event{
			ID:        m.ID,
			GameID:    m.GameID,
			PlayerID:  m.PlayerID,
			Type:      m.Type,
			Amount:    m.Amount,
			Metadata:  m.Metadata,
			CreatedAt: m.CreatedAt,
			CreatedBy: m.CreatedBy,
		})
	}
	return events, nil
}
