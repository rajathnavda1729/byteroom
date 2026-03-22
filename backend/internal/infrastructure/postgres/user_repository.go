package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/byteroom/backend/internal/domain/user"
)

// UserRepository is a PostgreSQL implementation of user.Repository.
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	const q = `
		INSERT INTO users (id, username, display_name, avatar_url, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, q,
		u.ID, u.Username, u.DisplayName, u.AvatarURL, u.PasswordHash, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return user.ErrDuplicateUsername
		}
		return fmt.Errorf("saving user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	const q = `
		SELECT id, username, display_name, avatar_url, password_hash, created_at, updated_at
		FROM users WHERE id = $1`

	u := &user.User{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.Username, &u.DisplayName, &u.AvatarURL, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by id: %w", err)
	}
	return u, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	const q = `
		SELECT id, username, display_name, avatar_url, password_hash, created_at, updated_at
		FROM users WHERE username = $1`

	u := &user.User{}
	err := r.db.QueryRowContext(ctx, q, username).Scan(
		&u.ID, &u.Username, &u.DisplayName, &u.AvatarURL, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by username: %w", err)
	}
	return u, nil
}

func (r *UserRepository) Search(ctx context.Context, query string, excludeID string, limit int) ([]*user.User, error) {
	const q = `
		SELECT id, username, display_name, avatar_url, password_hash, created_at, updated_at
		FROM users
		WHERE (username ILIKE $1 OR display_name ILIKE $1)
		  AND id != $2
		ORDER BY username
		LIMIT $3`

	rows, err := r.db.QueryContext(ctx, q, "%"+query+"%", excludeID, limit)
	if err != nil {
		return nil, fmt.Errorf("searching users: %w", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		u := &user.User{}
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.AvatarURL,
			&u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	const q = `
		UPDATE users SET display_name=$1, avatar_url=$2, updated_at=$3
		WHERE id=$4`

	res, err := r.db.ExecContext(ctx, q, u.DisplayName, u.AvatarURL, u.UpdatedAt, u.ID)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return user.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM users WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return user.ErrUserNotFound
	}
	return nil
}
