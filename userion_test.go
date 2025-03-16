package userion

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// createTestUser creates a test user for testing
func createTestUser(t *testing.T, userManager UserManager) *User {
	user := &User{
		ID:       uuid.New(),
		Name:     "Test User",
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Phone:    "1234567890",
		Enabled:  true,
		Status:   UserStatusActive,
		Data: map[string]interface{}{
			"testKey": "testValue",
		},
	}

	err := userManager.CreateUser(user)
	require.NoError(t, err, "Failed to create test user")

	return user
}
