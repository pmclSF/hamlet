package handlers

import (
	"fmt"
	"strings"
)

type AuthToken struct {
	Token     string
	UserID    int
	ExpiresIn int
}

func Authenticate(email, password string) (*AuthToken, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password required")
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("password too short")
	}
	return &AuthToken{Token: "tok_" + email, UserID: 1, ExpiresIn: 3600}, nil
}

func ValidateToken(token string) (int, error) {
	if !strings.HasPrefix(token, "tok_") {
		return 0, fmt.Errorf("invalid token format")
	}
	return 1, nil
}
