package userion

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	// UserStatusActive represents an active user
	UserStatusActive UserStatus = "active"
	// UserStatusSuspended represents a suspended user
	UserStatusSuspended UserStatus = "suspended"
	// UserStatusLocked represents a locked user
	UserStatusLocked UserStatus = "locked"
	// UserStatusInactive represents an inactive user
	UserStatusInactive UserStatus = "inactive"
)

// Default values for new users
const (
	DefaultUserStatus = UserStatusInactive
)

// Common errors returned by the UserManager
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
)

// User represents the business model for user operations
type User struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	Password  string                 `json:"password,omitempty"`
	Salt      string                 `json:"-"` // Salt is never exposed in JSON
	Phone     string                 `json:"phone"`
	CreatedAt time.Time              `json:"created_at"`
	Enabled   bool                   `json:"enabled"`
	Status    UserStatus             `json:"status"`
	Data      map[string]interface{} `json:"data,omitempty"` // JSON data for custom extensions
}

// UserManager defines the interface for managing users
type UserManager interface {
	AutoMigrate() error
	CreateUser(user *User) error
	GetUserByID(id string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUserByID(id string, updatedData map[string]interface{}) error
	UpdateUserByUsername(username string, updatedData map[string]interface{}) error
	UpdateUserByEmail(email string, updatedData map[string]interface{}) error
	DeleteUserByID(id string) error
	DeleteUserByUsername(username string) error
	DeleteUserByEmail(email string) error
	VerifyPasswordByUsername(username, password string) error
	VerifyPasswordByEmail(email, password string) error
	VerifyPasswordByID(id string, password string) error
	ListUsers(limit, offset int, filters map[string]interface{}, orderBy string, desc bool) ([]User, error)
	EnableUserByID(id string) error
	DisableUserByID(id string) error
	SetUserStatusByID(id string, status UserStatus) error
	SetUserStatusByUsername(username string, status UserStatus) error
	SetUserStatusByEmail(email string, status UserStatus) error
}
