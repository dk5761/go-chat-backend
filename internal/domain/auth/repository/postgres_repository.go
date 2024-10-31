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

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (id, email, password_hash, token_version, last_login)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.PasswordHash, user.TokenVersion, user.LastLogin)
	if err != nil {
		return common.ErrConflict // e.g., if a duplicate email is inserted
	}
	return nil
}

func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, email, password_hash, token_version, last_login
        FROM users
        WHERE email = $1
    `
	row := r.db.QueryRow(ctx, query, email)

	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.TokenVersion, &user.LastLogin)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) UpdateTokenVersion(ctx context.Context, userID uuid.UUID, tokenVersion int) error {
	query := `
        UPDATE users
        SET token_version = $1
        WHERE id = $2
    `
	cmdTag, err := r.db.Exec(ctx, query, tokenVersion, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return common.ErrNotFound // No user with this ID
	}
	return nil
}

func (r *postgresUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLogin time.Time) error {
	query := `
        UPDATE users
        SET last_login = $1
        WHERE id = $2
    `
	cmdTag, err := r.db.Exec(ctx, query, lastLogin, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return common.ErrNotFound // No user with this ID
	}
	return nil
}

func (r *postgresUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	query := `
        SELECT id, email, password_hash, token_version, last_login
        FROM users
        WHERE id = $1
    `
	row := r.db.QueryRow(ctx, query, userID)

	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.TokenVersion, &user.LastLogin)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
