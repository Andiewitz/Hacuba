package tokens

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAccessRoundTrip(t *testing.T) {
	secret := []byte("test-secret-must-be-32-bytes-minimum!!")
	userID := uuid.MustParse("0193a1b2-c3d4-7e5f-8901-23456789abcd")
	csrfHash := HashToken("csrf-raw-value")

	raw, err := IssueAccess(secret, userID, csrfHash, 15*time.Minute)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	gotID, gotHash, err := VerifyAccess(secret, raw)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if gotID != userID {
		t.Errorf("user mismatch: got %s want %s", gotID, userID)
	}
	if gotHash != csrfHash {
		t.Errorf("csrf binding lost")
	}
}

func TestAccessRejectsWrongSecret(t *testing.T) {
	good := []byte("test-secret-must-be-32-bytes-minimum!!")
	bad := []byte("wrong-secret-must-be-32-bytes-minimum!")
	raw, err := IssueAccess(good, uuid.Must(uuid.NewV7()), HashToken("x"), time.Minute)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, _, err := VerifyAccess(bad, raw); err == nil {
		t.Error("expected verification with wrong secret to fail")
	}
}

func TestOpaqueTokenHashes(t *testing.T) {
	raw, hash, err := NewOpaqueToken()
	if err != nil {
		t.Fatalf("new token: %v", err)
	}
	if raw == hash {
		t.Error("raw token must differ from stored hash")
	}
	if !EqualHash(raw, hash) {
		t.Error("EqualHash should accept the original token")
	}
	if EqualHash(raw+"tampered", hash) {
		t.Error("EqualHash should reject tampered tokens")
	}
}
