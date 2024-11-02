package repository

import (
	"context"
	"errors"
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
	// Initialize LastLogin to current time for new users
	user.LastLogin = time.Now()
	user.LastLoginToken = time.Now()

	query := `
        INSERT INTO users (id, email, username, password_hash, created_at, updated_at, last_login, last_login_token)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt, user.LastLogin, user.LastLoginToken)
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

// GetUserByUsername retrieves a user by username, including the updated timestamp fields
func (r *postgresUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
        SELECT id, email, username, password_hash, created_at, updated_at, last_login, last_login_token
        FROM users
        WHERE username = $2
    `
	row := r.db.QueryRow(ctx, query, username)

	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.LastLoginToken)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) GetUsers(ctx context.Context, q string, limit, offset int) ([]*models.User, int, error) {
	query := `
        WITH users_with_count AS (
            SELECT 
                id, username, email, created_at, updated_at,
                COUNT(*) OVER() AS total_count
            FROM users
            WHERE username ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%'
            ORDER BY created_at DESC
            LIMIT $2 OFFSET $3
        )
        SELECT id, username, email, created_at, updated_at, total_count FROM users_with_count;
    `

	rows, err := r.db.Query(ctx, query, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	var totalItems int
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt, &totalItems); err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, totalItems, nil
}

// UpdateLastLogin updates the last login time and last login token for the user
func (r *postgresUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLogin, lastLoginToken time.Time) error {
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

func (r *postgresUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
        UPDATE users
        SET email = $1, updated_at = $2
        WHERE id = $3
    `

	_, err := r.db.Exec(ctx, query, user.Email, time.Now(), user.ID)
	return err
}

// DeleteUser deletes a user from the database by their user ID.
func (r *postgresUserRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	cmdTag, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	// Check if any rows were actually deleted; if none, return an error indicating user not found
	if cmdTag.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}
