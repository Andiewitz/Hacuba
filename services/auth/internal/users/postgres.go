package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore is the sole Store implementation. It owns the only
// connection pool to auth-db — no other service gets credentials or
// network access to this database.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore connects to Postgres. Callers pass cfg.DatabaseURL,
// which resolves to auth-db:5432 on the internal Docker network.
func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

// Close releases the pool (deferred in main.go).
func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) CreateUser(ctx context.Context, user User) (*User, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, created_at`,
		user.ID, user.Email, user.PasswordHash,
	)
	var created User
	if err := row.Scan(&created.ID, &created.Email, &created.PasswordHash, &created.CreatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &created, nil
}

func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`,
		NormalizeEmail(email),
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

func (s *PostgresStore) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (s *PostgresStore) CreateRefreshSession(ctx context.Context, sess RefreshSession) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO refresh_sessions (id, user_id, token_hash, csrf_hash, expires_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		sess.ID, sess.UserID, sess.TokenHash, sess.CSRFHash, sess.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create refresh session: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetRefreshSessionByHash(ctx context.Context, tokenHash string) (*RefreshSession, error) {
	var sess RefreshSession
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, csrf_hash, expires_at, created_at
		 FROM refresh_sessions WHERE token_hash = $1`,
		tokenHash,
	).Scan(&sess.ID, &sess.UserID, &sess.TokenHash, &sess.CSRFHash, &sess.ExpiresAt, &sess.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get refresh session: %w", err)
	}
	return &sess, nil
}

func (s *PostgresStore) DeleteRefreshSessionByHash(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM refresh_sessions WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete refresh session: %w", err)
	}
	return nil
}

func (s *PostgresStore) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM refresh_sessions WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete user sessions: %w", err)
	}
	return nil
}

// isUniqueViolation detects Postgres unique violations (SQLSTATE 23505)
// without leaking driver details to call sites.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
