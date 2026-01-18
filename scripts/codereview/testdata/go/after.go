package example

import (
	"context"
	"fmt"
)

// User represents a user in the system
type User struct {
	ID       int
	Name     string
	Email    string // Added field
	IsActive bool   // Added field
}

// Greeter interface for greeting
type Greeter interface {
	Greet() string
	GreetWithContext(ctx context.Context) string // Added method
}

// Config is a new type
type Config struct {
	Debug   bool
	Timeout int
}

const DefaultTimeout = 30

var greetingPrefix = "Hello"

// Hello returns a greeting message (signature changed)
func Hello(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	return fmt.Sprintf("Hello, %s!", name), nil
}

// (u *User) GetName returns the user's name
func (u *User) GetName() string {
	return u.Name
}

// (u *User) GetEmail is a new method
func (u *User) GetEmail() string {
	return u.Email
}

// NewGreeting is a new function
func NewGreeting(prefix, name string) string {
	return fmt.Sprintf("%s, %s!", prefix, name)
}
