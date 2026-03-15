package handlers

import "testing"

func TestAuthenticate(t *testing.T) {
	t.Parallel()
	tok, err := Authenticate("user@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestAuthenticate_ShortPassword(t *testing.T) {
	t.Parallel()
	_, err := Authenticate("user@example.com", "short")
	if err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestValidateToken(t *testing.T) {
	t.Parallel()
	uid, err := ValidateToken("tok_user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid != 1 {
		t.Errorf("uid = %d, want 1", uid)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	t.Parallel()
	_, err := ValidateToken("bad_token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
