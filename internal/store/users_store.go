package store

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	FirebaseUID *string   `json:"firebase_uid"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateUserInput struct {
	Email       string  `json:"email"`
	Name        string  `json:"name"`
	FirebaseUID *string `json:"firebase_uid"`
}

type UpsertUserFromFirebaseInput struct {
	FirebaseUID string
	Email       string
	Name        string
}

func (s *Store) GetUserByFirebaseUID(ctx context.Context, uid string) (*User, error) {
	const query = `
		SELECT id, email, name, firebase_uid, created_at, updated_at
		FROM app_user
		WHERE firebase_uid = $1`

	var user User
	err := s.DB.QueryRowContext(ctx, query, uid).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.FirebaseUID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by firebase_uid %s: %w", uid, err)
	}

	return &user, nil
}

func (s *Store) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
	const query = `
		INSERT INTO app_user (email, name, firebase_uid)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, firebase_uid, created_at, updated_at`

	var user User
	err := s.DB.QueryRowContext(ctx, query, input.Email, input.Name, input.FirebaseUID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.FirebaseUID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (s *Store) UpsertUserFromFirebase(ctx context.Context, input UpsertUserFromFirebaseInput) (*User, error) {
	name := firebaseName(input.Name, input.Email)

	const query = `
		INSERT INTO app_user (email, name, firebase_uid)
		VALUES ($1, $2, $3)
		ON CONFLICT (firebase_uid) DO UPDATE
		SET email = EXCLUDED.email,
			name = EXCLUDED.name,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, email, name, firebase_uid, created_at, updated_at`

	var user User
	err := s.DB.QueryRowContext(ctx, query, input.Email, name, input.FirebaseUID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.FirebaseUID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert user from firebase uid %s: %w", input.FirebaseUID, err)
	}

	return &user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (*User, error) {
	const query = `
		SELECT id, email, name, firebase_uid, created_at, updated_at
		FROM app_user
		WHERE id = $1`

	var user User
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.FirebaseUID,
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
		SELECT id, email, name, firebase_uid, created_at, updated_at
		FROM app_user
		WHERE email = $1`

	var user User
	err := s.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.FirebaseUID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by email %s: %w", email, err)
	}

	return &user, nil
}

func firebaseName(name string, email string) string {
	name = strings.TrimSpace(name)
	if name != "" {
		return name
	}

	email = strings.TrimSpace(email)
	localPart, _, found := strings.Cut(email, "@")
	if found && localPart != "" {
		return localPart
	}

	return "User"
}
