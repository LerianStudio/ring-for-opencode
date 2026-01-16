package example

import (
	"fmt"
	"strings"
)

// User represents a user in the system
type User struct {
	ID   int
	Name string
}

// Greeter interface for greeting
type Greeter interface {
	Greet() string
}

// Hello returns a greeting message
func Hello(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// (u *User) GetName returns the user's name
func (u *User) GetName() string {
	return u.Name
}

// FormatName formats a name with optional prefix
func FormatName(name string) string {
	return strings.Title(name)
}
