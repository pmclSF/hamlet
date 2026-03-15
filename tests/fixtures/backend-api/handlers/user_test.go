package handlers

import "testing"

func TestCreateUser(t *testing.T) {
	t.Parallel()
	u, err := CreateUser("Alice", "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Name != "Alice" {
		t.Errorf("name = %q, want Alice", u.Name)
	}
}

func TestCreateUser_EmptyName(t *testing.T) {
	t.Parallel()
	_, err := CreateUser("", "alice@example.com")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestGetUser(t *testing.T) {
	t.Parallel()
	u, err := GetUser(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != 1 {
		t.Errorf("id = %d, want 1", u.ID)
	}
}

func TestGetUser_InvalidID(t *testing.T) {
	t.Parallel()
	_, err := GetUser(0)
	if err == nil {
		t.Fatal("expected error for invalid id")
	}
}
