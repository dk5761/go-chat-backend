package auth

import (
	"context"
	"errors"
	"time"

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

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
        INSERT INTO users (id, email, password_hash, token_version, last_login)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.PasswordHash, user.TokenVersion, user.LastLogin)
	if err != nil {
		return err
	}
	return nil
}

func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
        SELECT id, email, password_hash, token_version, last_login
        FROM users
        WHERE email = $1
    `
	row := r.db.QueryRow(ctx, query, email)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.TokenVersion, &user.LastLogin)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("user not found")
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
		return errors.New("no rows were updated")
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
		return errors.New("no rows were updated")
	}
	return nil
}

func (r *postgresUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	query := `
        SELECT id, email, password_hash, token_version, last_login
        FROM users
        WHERE id = $1
    `
	row := r.db.QueryRow(ctx, query, userID)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.TokenVersion, &user.LastLogin)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}
