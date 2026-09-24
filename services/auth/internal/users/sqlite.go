package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// SQLiteStore is a persistent local-development implementation. Production
// continues to use Postgres and its migrations; this compact schema is kept
// separate so SQLite behaviour cannot silently become production behaviour.
type SQLiteStore struct {
	db *sql.DB
	mu sync.Mutex
}

func NewSQLiteStore(ctx context.Context, path string) (*SQLiteStore, error) {
	if path == "" {
		return nil, fmt.Errorf("SQLite path is required")
	}
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open development SQLite: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping development SQLite: %w", err)
	}
	store := &SQLiteStore{db: db}
	if err := store.initialize(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) initialize(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS dev_auth_users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('buyer', 'seller')),
			created_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS dev_auth_refresh_sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			csrf_hash TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_dev_auth_refresh_user ON dev_auth_refresh_sessions(user_id);
	`)
	if err != nil {
		return fmt.Errorf("initialize development SQLite: %w", err)
	}
	return nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

func (s *SQLiteStore) CreateUser(ctx context.Context, user User) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user.Email = NormalizeEmail(user.Email)
	user.Role = NormalizeRole(user.Role)
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO dev_auth_users (id, email, password_hash, role, created_at) VALUES (?, ?, ?, ?, ?)`,
		user.ID.String(), user.Email, user.PasswordHash, user.Role, formatTime(user.CreatedAt),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create development user: %w", err)
	}
	return &user, nil
}

func (s *SQLiteStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.readUser(ctx, `SELECT id, email, password_hash, role, created_at FROM dev_auth_users WHERE email=?`, NormalizeEmail(email))
}

func (s *SQLiteStore) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.readUser(ctx, `SELECT id, email, password_hash, role, created_at FROM dev_auth_users WHERE id=?`, id.String())
}

func (s *SQLiteStore) BecomeSeller(ctx context.Context, id uuid.UUID) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result, err := s.db.ExecContext(ctx, `UPDATE dev_auth_users SET role='seller' WHERE id=?`, id.String())
	if err != nil {
		return nil, fmt.Errorf("become development seller: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, ErrNotFound
	}
	return s.readUser(ctx, `SELECT id, email, password_hash, role, created_at FROM dev_auth_users WHERE id=?`, id.String())
}

func (s *SQLiteStore) CreateRefreshSession(ctx context.Context, session RefreshSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO dev_auth_refresh_sessions (id, user_id, token_hash, csrf_hash, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		session.ID.String(), session.UserID.String(), session.TokenHash, session.CSRFHash, formatTime(session.ExpiresAt), formatTime(session.CreatedAt))
	if err != nil {
		return fmt.Errorf("create development refresh session: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetRefreshSessionByHash(ctx context.Context, tokenHash string) (*RefreshSession, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, user_id, token_hash, csrf_hash, expires_at, created_at FROM dev_auth_refresh_sessions WHERE token_hash=?`, tokenHash)
	return scanSession(row)
}

func (s *SQLiteStore) DeleteRefreshSessionByHash(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM dev_auth_refresh_sessions WHERE token_hash=?`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete development refresh session: %w", err)
	}
	return nil
}

func (s *SQLiteStore) RotateRefreshSession(ctx context.Context, oldHash string, next RefreshSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin development refresh rotation: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `DELETE FROM dev_auth_refresh_sessions WHERE token_hash=?`, oldHash)
	if err != nil {
		return fmt.Errorf("consume development refresh session: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	if next.CreatedAt.IsZero() {
		next.CreatedAt = time.Now().UTC()
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO dev_auth_refresh_sessions (id, user_id, token_hash, csrf_hash, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		next.ID.String(), next.UserID.String(), next.TokenHash, next.CSRFHash, formatTime(next.ExpiresAt), formatTime(next.CreatedAt))
	if err != nil {
		return fmt.Errorf("create rotated development refresh session: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit development refresh rotation: %w", err)
	}
	return nil
}

func (s *SQLiteStore) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM dev_auth_refresh_sessions WHERE user_id=?`, userID.String())
	if err != nil {
		return fmt.Errorf("delete development user sessions: %w", err)
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func (s *SQLiteStore) readUser(ctx context.Context, query string, arg string) (*User, error) {
	row := s.db.QueryRowContext(ctx, query, arg)
	var id, created string
	var user User
	if err := row.Scan(&id, &user.Email, &user.PasswordHash, &user.Role, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read development user: %w", err)
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("decode development user ID: %w", err)
	}
	user.ID = parsedID
	user.CreatedAt, err = parseTime(created)
	if err != nil {
		return nil, fmt.Errorf("decode development user creation time: %w", err)
	}
	return &user, nil
}

func scanSession(row scanner) (*RefreshSession, error) {
	var id, userID, expiresAt, createdAt string
	var session RefreshSession
	if err := row.Scan(&id, &userID, &session.TokenHash, &session.CSRFHash, &expiresAt, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read development refresh session: %w", err)
	}
	var err error
	if session.ID, err = uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("decode development refresh session ID: %w", err)
	}
	if session.UserID, err = uuid.Parse(userID); err != nil {
		return nil, fmt.Errorf("decode development refresh user ID: %w", err)
	}
	if session.ExpiresAt, err = parseTime(expiresAt); err != nil {
		return nil, fmt.Errorf("decode development refresh expiry: %w", err)
	}
	if session.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, fmt.Errorf("decode development refresh creation time: %w", err)
	}
	return &session, nil
}

func formatTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func parseTime(value string) (time.Time, error) { return time.Parse(time.RFC3339Nano, value) }
