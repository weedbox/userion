package userion

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDBGorm creates a test database and returns a UserManager
func setupTestDBGorm(t *testing.T) (UserManager, *gorm.DB) {
	// Use SQLite in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to database")

	// Register cleanup function to close the database when test completes
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	})

	// Create a new UserManager with a random table name to ensure test isolation
	tableName := "users_test_" + uuid.New().String()[:8]
	userManager := NewGormUserManager(db, tableName)

	// Run migrations
	err = userManager.AutoMigrate()
	require.NoError(t, err, "Failed to migrate database")

	return userManager, db
}

// TestAutoMigrate_Gorm tests the AutoMigrate method
func TestAutoMigrate_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)

	// AutoMigrate is already called in setupTestDBGorm, so just test that it doesn't error when called again
	err := userManager.AutoMigrate()
	assert.NoError(t, err, "AutoMigrate should not error when called multiple times")
}

// TestCreateUser_Gorm tests the CreateUser method
func TestCreateUser_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)

	// Test creating a new user
	user := &User{
		Name:     "New User",
		Username: "newuser",
		Email:    "new@example.com",
		Password: "newpassword",
		Phone:    "9876543210",
		Enabled:  true,
		Status:   UserStatusActive,
		Data: map[string]interface{}{
			"preferences": map[string]interface{}{
				"theme": "dark",
			},
		},
	}

	err := userManager.CreateUser(user)
	assert.NoError(t, err, "CreateUser should not error with valid user")
	assert.NotEqual(t, uuid.Nil, user.ID, "User ID should be generated")
	assert.NotEmpty(t, user.Salt, "User salt should be generated")
	assert.NotEqual(t, "newpassword", user.Password, "User password should be hashed")

	// Test creating a user with existing username
	duplicateUser := &User{
		Name:     "Duplicate User",
		Username: "newuser", // Same as above
		Email:    "different@example.com",
		Password: "duplicatepassword",
		Phone:    "5555555555",
	}

	err = userManager.CreateUser(duplicateUser)
	assert.Equal(t, ErrUserAlreadyExists, err, "CreateUser should error with duplicate username")

	// Test creating a user with existing email
	duplicateUser = &User{
		Name:     "Duplicate User",
		Username: "differentuser",
		Email:    "new@example.com", // Same as above
		Password: "duplicatepassword",
		Phone:    "5555555555",
	}

	err = userManager.CreateUser(duplicateUser)
	assert.Equal(t, ErrUserAlreadyExists, err, "CreateUser should error with duplicate email")
}

// TestGetUserByID_Gorm tests the GetUserByID method
func TestGetUserByID_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := createTestUser(t, userManager)

	// Test getting a user by ID
	retrievedUser, err := userManager.GetUserByID(user.ID.String())
	assert.NoError(t, err, "GetUserByID should not error with valid ID")
	assert.Equal(t, user.ID, retrievedUser.ID, "Retrieved user ID should match")
	assert.Equal(t, user.Username, retrievedUser.Username, "Retrieved user username should match")
	assert.Equal(t, user.Email, retrievedUser.Email, "Retrieved user email should match")
	assert.Equal(t, "testValue", retrievedUser.Data["testKey"], "Retrieved user data should match")

	// Test getting a user with invalid ID
	invalidID := uuid.New().String()
	_, err = userManager.GetUserByID(invalidID)
	assert.Equal(t, ErrUserNotFound, err, "GetUserByID should return ErrUserNotFound with invalid ID")
}

// TestGetUserByUsername_Gorm tests the GetUserByUsername method
func TestGetUserByUsername_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := createTestUser(t, userManager)

	// Test getting a user by username
	retrievedUser, err := userManager.GetUserByUsername(user.Username)
	assert.NoError(t, err, "GetUserByUsername should not error with valid username")
	assert.Equal(t, user.ID, retrievedUser.ID, "Retrieved user ID should match")
	assert.Equal(t, user.Username, retrievedUser.Username, "Retrieved user username should match")

	// Test getting a user with invalid username
	_, err = userManager.GetUserByUsername("invalidusername")
	assert.Equal(t, ErrUserNotFound, err, "GetUserByUsername should return ErrUserNotFound with invalid username")
}

// TestGetUserByEmail_Gorm tests the GetUserByEmail method
func TestGetUserByEmail_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := createTestUser(t, userManager)

	// Test getting a user by email
	retrievedUser, err := userManager.GetUserByEmail(user.Email)
	assert.NoError(t, err, "GetUserByEmail should not error with valid email")
	assert.Equal(t, user.ID, retrievedUser.ID, "Retrieved user ID should match")
	assert.Equal(t, user.Email, retrievedUser.Email, "Retrieved user email should match")

	// Test getting a user with invalid email
	_, err = userManager.GetUserByEmail("invalid@example.com")
	assert.Equal(t, ErrUserNotFound, err, "GetUserByEmail should return ErrUserNotFound with invalid email")
}

// TestUpdateUserByID_Gorm tests the UpdateUserByID method
func TestUpdateUserByID_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := createTestUser(t, userManager)

	// Test updating user fields
	updatedData := map[string]interface{}{
		"Name":  "Updated Name",
		"Phone": "5555555555",
		"Data": map[string]interface{}{
			"updatedKey": "updatedValue",
		},
	}

	err := userManager.UpdateUserByID(user.ID.String(), updatedData)
	assert.NoError(t, err, "UpdateUserByID should not error with valid ID and data")

	// Verify changes
	updatedUser, err := userManager.GetUserByID(user.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedUser.Name, "Name should be updated")
	assert.Equal(t, "5555555555", updatedUser.Phone, "Phone should be updated")
	assert.Equal(t, "updatedValue", updatedUser.Data["updatedKey"], "Data should be updated")

	// Test updating with invalid ID
	invalidID := uuid.New().String()
	err = userManager.UpdateUserByID(invalidID, updatedData)
	assert.Equal(t, ErrUserNotFound, err, "UpdateUserByID should return ErrUserNotFound with invalid ID")

	// Test updating password
	passwordData := map[string]interface{}{
		"Password": "newpassword",
	}

	oldPassword := updatedUser.Password
	oldSalt := updatedUser.Salt

	err = userManager.UpdateUserByID(user.ID.String(), passwordData)
	assert.NoError(t, err, "UpdateUserByID should not error when updating password")

	// Verify password change
	updatedUser, err = userManager.GetUserByID(user.ID.String())
	assert.NoError(t, err)
	assert.NotEqual(t, oldPassword, updatedUser.Password, "Password should be updated")
	assert.NotEqual(t, oldSalt, updatedUser.Salt, "Salt should be updated")

	// Verify password works
	err = userManager.VerifyPasswordByID(user.ID.String(), "newpassword")
	assert.NoError(t, err, "New password should verify correctly")
}

// TestUpdateUserByUsername_Gorm tests the UpdateUserByUsername method
func TestUpdateUserByUsername_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := createTestUser(t, userManager)

	// Test updating user fields
	updatedData := map[string]interface{}{
		"Name": "Updated By Username",
	}

	err := userManager.UpdateUserByUsername(user.Username, updatedData)
	assert.NoError(t, err, "UpdateUserByUsername should not error with valid username and data")

	// Verify changes
	updatedUser, err := userManager.GetUserByID(user.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, "Updated By Username", updatedUser.Name, "Name should be updated")

	// Test updating with invalid username
	err = userManager.UpdateUserByUsername("invalidusername", updatedData)
	assert.Equal(t, ErrUserNotFound, err, "UpdateUserByUsername should return ErrUserNotFound with invalid username")
}

// TestUpdateUserByEmail_Gorm tests the UpdateUserByEmail method
func TestUpdateUserByEmail_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := createTestUser(t, userManager)

	// Test updating user fields
	updatedData := map[string]interface{}{
		"Name": "Updated By Email",
	}

	err := userManager.UpdateUserByEmail(user.Email, updatedData)
	assert.NoError(t, err, "UpdateUserByEmail should not error with valid email and data")

	// Verify changes
	updatedUser, err := userManager.GetUserByID(user.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, "Updated By Email", updatedUser.Name, "Name should be updated")

	// Test updating with invalid email
	err = userManager.UpdateUserByEmail("invalid@example.com", updatedData)
	assert.Equal(t, ErrUserNotFound, err, "UpdateUserByEmail should return ErrUserNotFound with invalid email")
}

// TestDeleteUserByID_Gorm tests the DeleteUserByID method
func TestDeleteUserByID_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := createTestUser(t, userManager)

	// Test deleting a user by ID
	err := userManager.DeleteUserByID(user.ID.String())
	assert.NoError(t, err, "DeleteUserByID should not error with valid ID")

	// Verify user is deleted
	_, err = userManager.GetUserByID(user.ID.String())
	assert.Equal(t, ErrUserNotFound, err, "User should be deleted")

	// Test deleting a user with invalid ID
	invalidID := uuid.New().String()
	err = userManager.DeleteUserByID(invalidID)
	assert.Equal(t, ErrUserNotFound, err, "DeleteUserByID should return ErrUserNotFound with invalid ID")
}

// TestDeleteUserByUsername_Gorm tests the DeleteUserByUsername method
func TestDeleteUserByUsername_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := createTestUser(t, userManager)

	// Test deleting a user by username
	err := userManager.DeleteUserByUsername(user.Username)
	assert.NoError(t, err, "DeleteUserByUsername should not error with valid username")

	// Verify user is deleted
	_, err = userManager.GetUserByUsername(user.Username)
	assert.Equal(t, ErrUserNotFound, err, "User should be deleted")

	// Test deleting a user with invalid username
	err = userManager.DeleteUserByUsername("invalidusername")
	assert.Equal(t, ErrUserNotFound, err, "DeleteUserByUsername should return ErrUserNotFound with invalid username")
}

// TestDeleteUserByEmail_Gorm tests the DeleteUserByEmail method
func TestDeleteUserByEmail_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := createTestUser(t, userManager)

	// Test deleting a user by email
	err := userManager.DeleteUserByEmail(user.Email)
	assert.NoError(t, err, "DeleteUserByEmail should not error with valid email")

	// Verify user is deleted
	_, err = userManager.GetUserByEmail(user.Email)
	assert.Equal(t, ErrUserNotFound, err, "User should be deleted")

	// Test deleting a user with invalid email
	err = userManager.DeleteUserByEmail("invalid@example.com")
	assert.Equal(t, ErrUserNotFound, err, "DeleteUserByEmail should return ErrUserNotFound with invalid email")
}

// TestVerifyPasswordByUsername_Gorm tests the VerifyPasswordByUsername method
func TestVerifyPasswordByUsername_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := &User{
		Name:     "Password Test",
		Username: "passwordtest",
		Email:    "password@test.com",
		Password: "correct_password",
		Phone:    "1231231234",
	}

	err := userManager.CreateUser(user)
	require.NoError(t, err)

	// Test with correct password
	err = userManager.VerifyPasswordByUsername("passwordtest", "correct_password")
	assert.NoError(t, err, "VerifyPasswordByUsername should not error with correct password")

	// Test with incorrect password
	err = userManager.VerifyPasswordByUsername("passwordtest", "wrong_password")
	assert.Equal(t, ErrInvalidPassword, err, "VerifyPasswordByUsername should return ErrInvalidPassword with incorrect password")

	// Test with non-existent username
	err = userManager.VerifyPasswordByUsername("nonexistentuser", "any_password")
	assert.Equal(t, ErrUserNotFound, err, "VerifyPasswordByUsername should return ErrUserNotFound with non-existent username")
}

// TestVerifyPasswordByEmail_Gorm tests the VerifyPasswordByEmail method
func TestVerifyPasswordByEmail_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := &User{
		Name:     "Password Test",
		Username: "passwordtest",
		Email:    "password@test.com",
		Password: "correct_password",
		Phone:    "1231231234",
	}

	err := userManager.CreateUser(user)
	require.NoError(t, err)

	// Test with correct password
	err = userManager.VerifyPasswordByEmail("password@test.com", "correct_password")
	assert.NoError(t, err, "VerifyPasswordByEmail should not error with correct password")

	// Test with incorrect password
	err = userManager.VerifyPasswordByEmail("password@test.com", "wrong_password")
	assert.Equal(t, ErrInvalidPassword, err, "VerifyPasswordByEmail should return ErrInvalidPassword with incorrect password")

	// Test with non-existent email
	err = userManager.VerifyPasswordByEmail("nonexistent@test.com", "any_password")
	assert.Equal(t, ErrUserNotFound, err, "VerifyPasswordByEmail should return ErrUserNotFound with non-existent email")
}

// TestVerifyPasswordByID_Gorm tests the VerifyPasswordByID method
func TestVerifyPasswordByID_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := &User{
		ID:       uuid.New(),
		Name:     "Password Test",
		Username: "passwordtest",
		Email:    "password@test.com",
		Password: "correct_password",
		Phone:    "1231231234",
	}

	err := userManager.CreateUser(user)
	require.NoError(t, err)

	// Test with correct password
	err = userManager.VerifyPasswordByID(user.ID.String(), "correct_password")
	assert.NoError(t, err, "VerifyPasswordByID should not error with correct password")

	// Test with incorrect password
	err = userManager.VerifyPasswordByID(user.ID.String(), "wrong_password")
	assert.Equal(t, ErrInvalidPassword, err, "VerifyPasswordByID should return ErrInvalidPassword with incorrect password")

	// Test with non-existent ID
	nonExistentID := uuid.New().String()
	err = userManager.VerifyPasswordByID(nonExistentID, "any_password")
	assert.Equal(t, ErrUserNotFound, err, "VerifyPasswordByID should return ErrUserNotFound with non-existent ID")
}

// TestListUsers_Gorm tests the ListUsers method
func TestListUsers_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)

	// Create multiple users for testing
	for i := 0; i < 5; i++ {
		user := &User{
			Name:     "List User",
			Username: "listuser" + time.Now().String(),
			Email:    "list" + time.Now().String() + "@example.com",
			Password: "listpassword",
			Phone:    fmt.Sprintf("123456789%d", i),
			Enabled:  true,
			Status:   UserStatusActive,
		}
		err := userManager.CreateUser(user)
		require.NoError(t, err)

		// Need a small delay to ensure unique timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Test listing all users
	users, err := userManager.ListUsers(10, 0, nil, "", false)
	assert.NoError(t, err, "ListUsers should not error")
	assert.Equal(t, 5, len(users), "ListUsers should return all users")

	// Test pagination
	users, err = userManager.ListUsers(3, 0, nil, "", false)
	assert.NoError(t, err, "ListUsers with limit should not error")
	assert.Equal(t, 3, len(users), "ListUsers should respect limit")

	// Test with offset
	users, err = userManager.ListUsers(3, 3, nil, "", false)
	assert.NoError(t, err, "ListUsers with offset should not error")
	assert.Equal(t, 2, len(users), "ListUsers should respect offset")

	// Test with filters
	filterUser := &User{
		Name:     "Filtered User",
		Username: "filtereduser",
		Email:    "filtered@example.com",
		Password: "filteredpassword",
		Phone:    "9999999999",
		Enabled:  true,
		Status:   UserStatusInactive,
	}
	err = userManager.CreateUser(filterUser)
	require.NoError(t, err)

	// Filter by status
	users, err = userManager.ListUsers(10, 0, map[string]interface{}{"status": UserStatusInactive}, "", false)
	assert.NoError(t, err, "ListUsers with filters should not error")
	assert.Equal(t, 1, len(users), "ListUsers should filter users")
	assert.Equal(t, UserStatusInactive, users[0].Status, "Filter should return users with inactive status")

	// Test with ordering
	users, err = userManager.ListUsers(10, 0, nil, "created_at", true)
	assert.NoError(t, err, "ListUsers with ordering should not error")
	assert.Equal(t, 6, len(users), "ListUsers should return all users")

	// The list should be in descending order of creation time
	for i := 0; i < len(users)-1; i++ {
		assert.True(t, users[i].CreatedAt.After(users[i+1].CreatedAt) || users[i].CreatedAt.Equal(users[i+1].CreatedAt),
			"ListUsers should order by created_at DESC")
	}
}

// TestEnableDisableUser_Gorm tests the EnableUserByID and DisableUserByID methods
func TestEnableDisableUser_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := createTestUser(t, userManager)

	// Test disabling a user
	err := userManager.DisableUserByID(user.ID.String())
	assert.NoError(t, err, "DisableUserByID should not error with valid ID")

	// Verify user is disabled
	disabledUser, err := userManager.GetUserByID(user.ID.String())
	assert.NoError(t, err)
	assert.False(t, disabledUser.Enabled, "User should be disabled")

	// Test enabling a user
	err = userManager.EnableUserByID(user.ID.String())
	assert.NoError(t, err, "EnableUserByID should not error with valid ID")

	// Verify user is enabled
	enabledUser, err := userManager.GetUserByID(user.ID.String())
	assert.NoError(t, err)
	assert.True(t, enabledUser.Enabled, "User should be enabled")

	// Test with invalid ID
	invalidID := uuid.New().String()
	err = userManager.EnableUserByID(invalidID)
	assert.Equal(t, ErrUserNotFound, err, "EnableUserByID should return ErrUserNotFound with invalid ID")

	err = userManager.DisableUserByID(invalidID)
	assert.Equal(t, ErrUserNotFound, err, "DisableUserByID should return ErrUserNotFound with invalid ID")
}

// TestSetUserStatus_Gorm tests the SetUserStatusByID, SetUserStatusByUsername, and SetUserStatusByEmail methods
func TestSetUserStatus_Gorm(t *testing.T) {
	userManager, _ := setupTestDBGorm(t)
	user := createTestUser(t, userManager)

	// Test setting status by ID
	err := userManager.SetUserStatusByID(user.ID.String(), UserStatusSuspended)
	assert.NoError(t, err, "SetUserStatusByID should not error with valid ID")

	// Verify status
	updatedUser, err := userManager.GetUserByID(user.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, UserStatusSuspended, updatedUser.Status, "User status should be updated")

	// Test setting status by username
	err = userManager.SetUserStatusByUsername(user.Username, UserStatusLocked)
	assert.NoError(t, err, "SetUserStatusByUsername should not error with valid username")

	// Verify status
	updatedUser, err = userManager.GetUserByID(user.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, UserStatusLocked, updatedUser.Status, "User status should be updated")

	// Test setting status by email
	err = userManager.SetUserStatusByEmail(user.Email, UserStatusInactive)
	assert.NoError(t, err, "SetUserStatusByEmail should not error with valid email")

	// Verify status
	updatedUser, err = userManager.GetUserByID(user.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, UserStatusInactive, updatedUser.Status, "User status should be updated")

	// Test with invalid identifiers
	invalidID := uuid.New().String()
	err = userManager.SetUserStatusByID(invalidID, UserStatusActive)
	assert.Equal(t, ErrUserNotFound, err, "SetUserStatusByID should return ErrUserNotFound with invalid ID")

	err = userManager.SetUserStatusByUsername("invalidusername", UserStatusActive)
	assert.Equal(t, ErrUserNotFound, err, "SetUserStatusByUsername should return ErrUserNotFound with invalid username")

	err = userManager.SetUserStatusByEmail("invalid@example.com", UserStatusActive)
	assert.Equal(t, ErrUserNotFound, err, "SetUserStatusByEmail should return ErrUserNotFound with invalid email")
}
