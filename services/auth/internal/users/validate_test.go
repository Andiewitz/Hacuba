package users

import "testing"

func TestValidateEmail(t *testing.T) {
	valid := []string{"user@example.com", "  USER@Example.COM  "}
	for _, e := range valid {
		if err := ValidateEmail(e); err != nil {
			t.Errorf("expected %q to be valid, got %v", e, err)
		}
	}
	invalid := []string{"", "not-an-email", "a@b", "@example.com", "user@.com"}
	for _, e := range invalid {
		if err := ValidateEmail(e); err == nil {
			t.Errorf("expected %q to be invalid", e)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("short1a"); err == nil {
		t.Error("expected short password to fail")
	}
	if err := ValidatePassword("longbutnonumbers"); err == nil {
		t.Error("expected password without numbers to fail")
	}
	if err := ValidatePassword("validpassword123"); err != nil {
		t.Errorf("expected valid password to pass, got %v", err)
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := NormalizeEmail("  USER@Example.COM "); got != "user@example.com" {
		t.Errorf("got %q", got)
	}
}
