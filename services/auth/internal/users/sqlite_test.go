package users

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSQLiteStorePersistsUserAndRefreshSession(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "auth.db")
	store, err := NewSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("open SQLite store: %v", err)
	}
	userID := uuid.Must(uuid.NewV7())
	if _, err := store.CreateUser(ctx, User{ID: userID, Email: "seller@example.test", PasswordHash: "hash"}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := store.BecomeSeller(ctx, userID); err != nil {
		t.Fatalf("become seller: %v", err)
	}
	session := RefreshSession{ID: uuid.Must(uuid.NewV7()), UserID: userID, TokenHash: "old", CSRFHash: "csrf", ExpiresAt: time.Now().Add(time.Hour)}
	if err := store.CreateRefreshSession(ctx, session); err != nil {
		t.Fatalf("create refresh session: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close SQLite store: %v", err)
	}

	reopened, err := NewSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("reopen SQLite store: %v", err)
	}
	defer reopened.Close()
	user, err := reopened.GetUserByEmail(ctx, "SELLER@example.test")
	if err != nil || user.Role != RoleSeller {
		t.Fatalf("reopened user = %#v, %v; want seller", user, err)
	}
	if _, err := reopened.GetRefreshSessionByHash(ctx, "old"); err != nil {
		t.Fatalf("reopened refresh session: %v", err)
	}
}
