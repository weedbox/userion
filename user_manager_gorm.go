package userion

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// GormUserModel represents the GORM-specific database model for users
type GormUserModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;"`
	Name      string         `gorm:"not null"`
	Username  string         `gorm:"unique;not null"`
	Email     string         `gorm:"unique;not null"`
	Password  string         `gorm:"not null"`
	Salt      string         `gorm:"not null"`
	Phone     string         `gorm:"unique;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	Enabled   bool           `gorm:"not null;default:true"`
	Status    UserStatus     `gorm:"type:varchar(10);not null;default:'inactive'"`
	Data      datatypes.JSON `gorm:"type:json;default:'{}'"` // JSON data for custom extensions
}

// ToUser converts a GormUserModel to a User business model
func (g *GormUserModel) ToUser() *User {
	// Convert JSON data to map
	data := make(map[string]interface{})
	if len(g.Data) > 0 {
		err := json.Unmarshal([]byte(g.Data), &data)
		if err != nil {
			// If there's an error, initialize an empty map
			data = make(map[string]interface{})
		}
	}

	return &User{
		ID:        g.ID,
		Name:      g.Name,
		Username:  g.Username,
		Email:     g.Email,
		Password:  g.Password,
		Salt:      g.Salt,
		Phone:     g.Phone,
		CreatedAt: g.CreatedAt,
		Enabled:   g.Enabled,
		Status:    g.Status,
		Data:      data,
	}
}

// NewGormUserModelFromUser converts a User business model to a GormUserModel
func NewGormUserModelFromUser(user *User) *GormUserModel {
	// Convert map to JSON data
	var jsonData datatypes.JSON
	if user.Data != nil {
		data, err := json.Marshal(user.Data)
		if err == nil {
			jsonData = datatypes.JSON(data)
		} else {
			// If there's an error, initialize empty JSON
			jsonData = datatypes.JSON("{}")
		}
	} else {
		jsonData = datatypes.JSON("{}")
	}

	return &GormUserModel{
		ID:        user.ID,
		Name:      user.Name,
		Username:  user.Username,
		Email:     user.Email,
		Password:  user.Password,
		Salt:      user.Salt,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
		Enabled:   user.Enabled,
		Status:    user.Status,
		Data:      jsonData,
	}
}

// GormUserManager is the concrete implementation using GORM
type GormUserManager struct {
	db        *gorm.DB
	tableName string
}

// NewGormUserManager initializes a new UserManager
func NewGormUserManager(db *gorm.DB, tableName string) UserManager {
	return &GormUserManager{
		db:        db,
		tableName: tableName,
	}
}

// AutoMigrate creates or updates the database schema for User model
func (m *GormUserManager) AutoMigrate() error {
	return m.db.Table(m.tableName).AutoMigrate(&GormUserModel{})
}

// VerifyPasswordByUsername verifies the password of a user by username
func (m *GormUserManager) VerifyPasswordByUsername(username, password string) error {
	var gormUser GormUserModel
	if err := m.db.Table(m.tableName).Select("password", "salt").Where("username = ?", username).First(&gormUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if gormUser.Password != HashPassword(password, gormUser.Salt) {
		return ErrInvalidPassword
	}

	return nil
}

// VerifyPasswordByEmail verifies the password of a user by email
func (m *GormUserManager) VerifyPasswordByEmail(email, password string) error {
	var gormUser GormUserModel
	if err := m.db.Table(m.tableName).Select("password", "salt").Where("email = ?", email).First(&gormUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if gormUser.Password != HashPassword(password, gormUser.Salt) {
		return ErrInvalidPassword
	}

	return nil
}

// VerifyPasswordByID verifies the password of a user by ID
func (m *GormUserManager) VerifyPasswordByID(id string, password string) error {
	var gormUser GormUserModel
	if err := m.db.Table(m.tableName).Select("password", "salt").Where("id = ?", id).First(&gormUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if gormUser.Password != HashPassword(password, gormUser.Salt) {
		return ErrInvalidPassword
	}

	return nil
}

// ListUsers retrieves a list of users with pagination, filtering, and sorting
func (m *GormUserManager) ListUsers(limit, offset int, filters map[string]interface{}, orderBy string, desc bool) ([]User, error) {
	var gormUsers []GormUserModel
	query := m.db.Table(m.tableName)

	// Apply filters if any
	if filters != nil {
		for key, value := range filters {
			query = query.Where(key+" = ?", value)
		}
	}

	// Apply ordering if specified
	if orderBy != "" {
		if desc {
			query = query.Order(orderBy + " DESC")
		} else {
			query = query.Order(orderBy)
		}
	}

	// Apply pagination
	query = query.Limit(limit).Offset(offset)

	// Execute the query
	if err := query.Find(&gormUsers).Error; err != nil {
		return nil, err
	}

	// Convert GormUserModel slice to User slice
	users := make([]User, len(gormUsers))
	for i, gormUser := range gormUsers {
		user := gormUser.ToUser()
		users[i] = *user
	}

	return users, nil
}

// CreateUser creates a new user
func (m *GormUserManager) CreateUser(user *User) error {
	// Generate UUID for user ID
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Create a new user with default values if not specified
	if user.Status == "" {
		user.Status = DefaultUserStatus
	}

	// Convert User to GormUserModel
	gormUser := NewGormUserModelFromUser(user)

	// Check if user already exists with the same username or email
	var count int64
	m.db.Table(m.tableName).Where("username = ? OR email = ?", user.Username, user.Email).Count(&count)
	if count > 0 {
		return ErrUserAlreadyExists
	}

	// Generate a random salt if not provided
	if user.Salt == "" {
		salt, err := GenerateSalt()
		if err != nil {
			return err
		}
		user.Salt = salt
		gormUser.Salt = salt
	}

	// Hash the password if it's provided in plain text
	if user.Password != "" && len(user.Password) < 64 {
		hashedPassword := HashPassword(user.Password, user.Salt)
		user.Password = hashedPassword
		gormUser.Password = hashedPassword
	}

	// Create the user
	if err := m.db.Table(m.tableName).Create(gormUser).Error; err != nil {
		return err
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (m *GormUserManager) GetUserByID(id string) (*User, error) {
	var gormUser GormUserModel
	if err := m.db.Table(m.tableName).Where("id = ?", id).First(&gormUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return gormUser.ToUser(), nil
}

// GetUserByUsername retrieves a user by username
func (m *GormUserManager) GetUserByUsername(username string) (*User, error) {
	var gormUser GormUserModel
	if err := m.db.Table(m.tableName).Where("username = ?", username).First(&gormUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return gormUser.ToUser(), nil
}

// GetUserByEmail retrieves a user by email
func (m *GormUserManager) GetUserByEmail(email string) (*User, error) {
	var gormUser GormUserModel
	if err := m.db.Table(m.tableName).Where("email = ?", email).First(&gormUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return gormUser.ToUser(), nil
}

// UpdateUserByID updates user fields by ID
func (m *GormUserManager) UpdateUserByID(id string, updatedData map[string]interface{}) error {
	// Check if updating password and handle hash
	if password, ok := updatedData["Password"].(string); ok && len(password) < 64 {
		// Generate a new salt
		newSalt, err := GenerateSalt()
		if err != nil {
			return err
		}

		// Hash the password with the new salt
		updatedData["Password"] = HashPassword(password, newSalt)
		updatedData["Salt"] = newSalt
	}

	// Check if updating Data field (which is map[string]interface{} in User but JSON in GormUserModel)
	if dataMap, ok := updatedData["Data"].(map[string]interface{}); ok {
		// Convert map to JSON
		jsonData, err := json.Marshal(dataMap)
		if err == nil {
			updatedData["Data"] = datatypes.JSON(jsonData)
		} else {
			delete(updatedData, "Data") // Remove Data key if conversion fails
		}
	}

	result := m.db.Table(m.tableName).Where("id = ?", id).Updates(updatedData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdateUserByUsername updates user fields by username
func (m *GormUserManager) UpdateUserByUsername(username string, updatedData map[string]interface{}) error {
	// Check if updating password and handle hash
	if password, ok := updatedData["Password"].(string); ok && len(password) < 64 {
		// Generate a new salt
		newSalt, err := GenerateSalt()
		if err != nil {
			return err
		}

		// Hash the password with the new salt
		updatedData["Password"] = HashPassword(password, newSalt)
		updatedData["Salt"] = newSalt
	}

	// Check if updating Data field (which is map[string]interface{} in User but JSON in GormUserModel)
	if dataMap, ok := updatedData["Data"].(map[string]interface{}); ok {
		// Convert map to JSON
		jsonData, err := json.Marshal(dataMap)
		if err == nil {
			updatedData["Data"] = datatypes.JSON(jsonData)
		} else {
			delete(updatedData, "Data") // Remove Data key if conversion fails
		}
	}

	result := m.db.Table(m.tableName).Where("username = ?", username).Updates(updatedData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdateUserByEmail updates user fields by email
func (m *GormUserManager) UpdateUserByEmail(email string, updatedData map[string]interface{}) error {
	// Check if updating password and handle hash
	if password, ok := updatedData["Password"].(string); ok && len(password) < 64 {
		// Generate a new salt
		newSalt, err := GenerateSalt()
		if err != nil {
			return err
		}

		// Hash the password with the new salt
		updatedData["Password"] = HashPassword(password, newSalt)
		updatedData["Salt"] = newSalt
	}

	// Check if updating Data field (which is map[string]interface{} in User but JSON in GormUserModel)
	if dataMap, ok := updatedData["Data"].(map[string]interface{}); ok {
		// Convert map to JSON
		jsonData, err := json.Marshal(dataMap)
		if err == nil {
			updatedData["Data"] = datatypes.JSON(jsonData)
		} else {
			delete(updatedData, "Data") // Remove Data key if conversion fails
		}
	}

	result := m.db.Table(m.tableName).Where("email = ?", email).Updates(updatedData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DeleteUserByID deletes a user by ID
func (m *GormUserManager) DeleteUserByID(id string) error {
	result := m.db.Table(m.tableName).Where("id = ?", id).Delete(&GormUserModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DeleteUserByUsername deletes a user by username
func (m *GormUserManager) DeleteUserByUsername(username string) error {
	result := m.db.Table(m.tableName).Where("username = ?", username).Delete(&GormUserModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DeleteUserByEmail deletes a user by email
func (m *GormUserManager) DeleteUserByEmail(email string) error {
	result := m.db.Table(m.tableName).Where("email = ?", email).Delete(&GormUserModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// EnableUserByID enables a user by ID
func (m *GormUserManager) EnableUserByID(id string) error {
	result := m.db.Table(m.tableName).Where("id = ?", id).Update("enabled", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DisableUserByID disables a user by ID
func (m *GormUserManager) DisableUserByID(id string) error {
	result := m.db.Table(m.tableName).Where("id = ?", id).Update("enabled", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SetUserStatusByID updates the user status by ID
func (m *GormUserManager) SetUserStatusByID(id string, status UserStatus) error {
	result := m.db.Table(m.tableName).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SetUserStatusByUsername updates the user status by username
func (m *GormUserManager) SetUserStatusByUsername(username string, status UserStatus) error {
	result := m.db.Table(m.tableName).Where("username = ?", username).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SetUserStatusByEmail updates the user status by email
func (m *GormUserManager) SetUserStatusByEmail(email string, status UserStatus) error {
	result := m.db.Table(m.tableName).Where("email = ?", email).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
