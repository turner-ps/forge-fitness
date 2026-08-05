package store

import (
	"context"
	"fmt"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserInput struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (s *Store) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
	const query = `
INSERT INTO app_user (email, name)
VALUES ($1, $2)
RETURNING id, email, name, created_at, updated_at`

	var user User
	err := s.DB.QueryRowContext(ctx, query, input.Email, input.Name).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (*User, error) {
	const query = `
SELECT id, email, name, created_at, updated_at
FROM app_user
WHERE id = $1`

	var user User
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by id %d: %w", id, err)
	}

	return &user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const query = `
SELECT id, email, name, created_at, updated_at
FROM app_user
WHERE email = $1`

	var user User
	err := s.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by email %s: %w", email, err)
	}

	return &user, nil
}
