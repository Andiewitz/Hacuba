package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hacuba/auth/internal/users"
)

func TestConfiguredStaffEmailReceivesStaffSession(t *testing.T) {
	cfg := testConfig()
	cfg.StaffEmails = map[string]struct{}{"staff@example.test": {}}
	store := users.NewMemoryStore()
	response := doJSON(t, Register(cfg, store), http.MethodPost, "/auth/register", map[string]string{"email": "staff@example.test", "password": "validpassword123"})
	if response.Code != http.StatusCreated {
		t.Fatalf("register = %d: %s", response.Code, response.Body.String())
	}
	var session SessionResponse
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	if session.User.Role != users.RoleStaff {
		t.Fatalf("role = %q", session.User.Role)
	}
}
