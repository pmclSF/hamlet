package handlers

import "fmt"

type User struct {
	ID    int
	Name  string
	Email string
}

func CreateUser(name, email string) (*User, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	return &User{ID: 1, Name: name, Email: email}, nil
}

func GetUser(id int) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id: %d", id)
	}
	return &User{ID: id, Name: "Test User", Email: "test@example.com"}, nil
}

func DeleteUser(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid user id: %d", id)
	}
	return nil
}
