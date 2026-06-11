package repository

import (
	"context"

	db "github.com/ChinthaVamsidharReddy/ainyx-backend-task/db/sqlc"
)

// UserRepository wraps SQLC Queries and exposes only the methods the service needs.
// This thin abstraction makes it easy to swap the underlying DB layer in tests.
type UserRepository struct {
	q *db.Queries
}

// New creates a UserRepository backed by the provided SQLC Queries instance.
func New(q *db.Queries) *UserRepository {
	return &UserRepository{q: q}
}

func (r *UserRepository) Create(ctx context.Context, params db.CreateUserParams) (db.User, error) {
	return r.q.CreateUser(ctx, params)
}

func (r *UserRepository) GetByID(ctx context.Context, id int32) (db.User, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *UserRepository) Update(ctx context.Context, params db.UpdateUserParams) (db.User, error) {
	return r.q.UpdateUser(ctx, params)
}

func (r *UserRepository) Delete(ctx context.Context, id int32) error {
	return r.q.DeleteUser(ctx, id)
}

func (r *UserRepository) List(ctx context.Context, params db.ListUsersParams) ([]db.User, error) {
	return r.q.ListUsers(ctx, params)
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountUsers(ctx)
}
