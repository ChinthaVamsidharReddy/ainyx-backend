package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	db "github.com/ChinthaVamsidharReddy/ainyx-backend-task/db/sqlc"
	"github.com/ChinthaVamsidharReddy/ainyx-backend-task/internal/models"
	"github.com/ChinthaVamsidharReddy/ainyx-backend-task/internal/repository"
	"go.uber.org/zap"
)

const dobLayout = "2006-01-02"

// UserService contains all business logic for the users resource.
type UserService struct {
	repo   *repository.UserRepository
	logger *zap.Logger
}

// New creates a UserService.
func New(repo *repository.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{repo: repo, logger: logger}
}

// CalculateAge returns the number of full years between dob and today.
// It correctly handles the case where the birthday has not yet occurred this year.
func CalculateAge(dob time.Time) int {
	today := time.Now()
	years := today.Year() - dob.Year()

	// If the birthday month/day hasn't arrived yet this year, subtract one year.
	if today.Month() < dob.Month() ||
		(today.Month() == dob.Month() && today.Day() < dob.Day()) {
		years--
	}
	return years
}

func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.UserResponse, error) {
	dob, err := time.Parse(dobLayout, req.DOB)
	if err != nil {
		return nil, fmt.Errorf("invalid dob format: %w", err)
	}

	user, err := s.repo.Create(ctx, db.CreateUserParams{
		Name: req.Name,
		Dob:  dob,
	})
	if err != nil {
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	s.logger.Info("user created", zap.Int32("id", user.ID), zap.String("name", user.Name))
	return toUserResponse(user), nil
}

func (s *UserService) GetUser(ctx context.Context, id int32) (*models.UserWithAgeResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		s.logger.Error("failed to get user", zap.Int32("id", id), zap.Error(err))
		return nil, fmt.Errorf("could not retrieve user: %w", err)
	}

	s.logger.Info("user fetched", zap.Int32("id", user.ID))
	return toUserWithAgeResponse(user), nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int32, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	dob, err := time.Parse(dobLayout, req.DOB)
	if err != nil {
		return nil, fmt.Errorf("invalid dob format: %w", err)
	}

	user, err := s.repo.Update(ctx, db.UpdateUserParams{
		ID:   id,
		Name: req.Name,
		Dob:  dob,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		s.logger.Error("failed to update user", zap.Int32("id", id), zap.Error(err))
		return nil, fmt.Errorf("could not update user: %w", err)
	}

	s.logger.Info("user updated", zap.Int32("id", user.ID))
	return toUserResponse(user), nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int32) error {
	// Verify existence before deleting so we can return 404 vs 204.
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("could not verify user existence: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete user", zap.Int32("id", id), zap.Error(err))
		return fmt.Errorf("could not delete user: %w", err)
	}

	s.logger.Info("user deleted", zap.Int32("id", id))
	return nil
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) (*models.PaginatedUsersResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := int32((page - 1) * pageSize)

	users, err := s.repo.List(ctx, db.ListUsersParams{
		Limit:  int32(pageSize),
		Offset: offset,
	})
	if err != nil {
		s.logger.Error("failed to list users", zap.Error(err))
		return nil, fmt.Errorf("could not list users: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		s.logger.Error("failed to count users", zap.Error(err))
		return nil, fmt.Errorf("could not count users: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	var data []models.UserWithAgeResponse
	for _, u := range users {
		data = append(data, *toUserWithAgeResponse(u))
	}

	s.logger.Info("users listed", zap.Int("count", len(data)), zap.Int("page", page))
	return &models.PaginatedUsersResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ── helpers ─────────────────────────────────────────────────────────────────

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("user not found")

func toUserResponse(u db.User) *models.UserResponse {
	return &models.UserResponse{
		ID:   u.ID,
		Name: u.Name,
		DOB:  u.Dob.Format(dobLayout),
	}
}

func toUserWithAgeResponse(u db.User) *models.UserWithAgeResponse {
	return &models.UserWithAgeResponse{
		ID:   u.ID,
		Name: u.Name,
		DOB:  u.Dob.Format(dobLayout),
		Age:  CalculateAge(u.Dob),
	}
}
