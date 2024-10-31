package repository

import (
	"context"
	"time"

	"github.com/dk5761/go-serv/internal/domain/auth/models"
	"github.com/dk5761/go-serv/internal/domain/common"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type postgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{db}
}

// CreateUser inserts a new user with created, updated, and last login token timestamps
func (r *postgresUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (id, email, password_hash, created_at, updated_at, last_login, last_login_token)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt, user.LastLogin, user.LastLoginToken)
	if err != nil {
		return common.ErrConflict // e.g., if a duplicate email is inserted
	}
	return nil
}

// GetUserByEmail retrieves a user by email, including the updated timestamp fields
func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, email, password_hash, created_at, updated_at, last_login, last_login_token
        FROM users
        WHERE email = $1
    `
	row := r.db.QueryRow(ctx, query, email)

	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.LastLoginToken)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// UpdateLastLogin updates the last login time and last login token for the user
func (r *postgresUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLogin time.Time, lastLoginToken time.Time) error {
	query := `
        UPDATE users
        SET last_login = $1, last_login_token = $2, updated_at = $3
        WHERE id = $4
    `
	cmdTag, err := r.db.Exec(ctx, query, lastLogin, lastLoginToken, time.Now(), userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return common.ErrNotFound // No user with this ID
	}
	return nil
}

// GetUserByID retrieves a user by ID, including the updated timestamp fields
func (r *postgresUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	query := `
        SELECT id, email, password_hash, created_at, updated_at, last_login, last_login_token
        FROM users
        WHERE id = $1
    `
	row := r.db.QueryRow(ctx, query, userID)

	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.LastLoginToken)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUserTimestamps updates the updated_at field of the user to reflect changes
func (r *postgresUserRepository) UpdateUserTimestamps(ctx context.Context, userID uuid.UUID, updatedAt time.Time) error {
	query := `
        UPDATE users
        SET updated_at = $1
        WHERE id = $2
    `
	cmdTag, err := r.db.Exec(ctx, query, updatedAt, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return common.ErrNotFound // No user with this ID
	}
	return nil
}
