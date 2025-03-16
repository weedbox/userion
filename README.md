# Userion

Userion is a Go library that provides a flexible and extensible user management system for Go applications. It offers a clean interface for common user management operations such as authentication, user creation, retrieval, updates, and more.

## Features

- Complete user lifecycle management (create, read, update, delete)
- Multiple lookup methods (by ID, username, or email)
- Password verification with salt-based hashing
- User status management (active, suspended, locked, inactive)
- Enable/disable user functionality
- Advanced user listing with pagination, filtering, and ordering
- Customizable user data through a flexible JSON data field
- GORM database integration

## Installation

```bash
go get github.com/weedbox/userion
```

## Usage

### Initialize User Manager

```go
import (
    "github.com/weedbox/userion"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func main() {
    // Connect to database
    db, err := gorm.Open(sqlite.Open("users.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    // Create user manager with custom table name (optional)
    userManager := userion.NewGormUserManager(db, "users")

    // Run auto migrations
    err = userManager.AutoMigrate()
    if err != nil {
        panic("failed to migrate database")
    }

    // Now you can use userManager to manage users
}
```

### Create a New User

```go
user := &userion.User{
    Name:     "John Doe",
    Username: "johndoe",
    Email:    "john@example.com",
    Password: "securepassword123",  // Will be automatically hashed
    Phone:    "1234567890",
    Enabled:  true,
    Status:   userion.UserStatusActive,
    Data: map[string]interface{}{
        "preferences": map[string]interface{}{
            "theme": "dark",
        },
    },
}

err := userManager.CreateUser(user)
if err != nil {
    // Handle error (e.g., user already exists)
}

// user.ID will be populated with a new UUID
```

### Get a User

```go
// Get by ID
user, err := userManager.GetUserByID("user-uuid-here")

// Get by username
user, err := userManager.GetUserByUsername("johndoe")

// Get by email
user, err := userManager.GetUserByEmail("john@example.com")
```

### Update a User

```go
updatedData := map[string]interface{}{
    "Name": "John Smith",
    "Phone": "9876543210",
    "Data": map[string]interface{}{
        "preferences": map[string]interface{}{
            "theme": "light",
        },
    },
}

err := userManager.UpdateUserByID("user-uuid-here", updatedData)
// or
err := userManager.UpdateUserByUsername("johndoe", updatedData)
// or
err := userManager.UpdateUserByEmail("john@example.com", updatedData)
```

### Verify Password

```go
err := userManager.VerifyPasswordByUsername("johndoe", "securepassword123")
// or
err := userManager.VerifyPasswordByEmail("john@example.com", "securepassword123")
// or
err := userManager.VerifyPasswordByID("user-uuid-here", "securepassword123")

if err == nil {
    // Password is correct
} else if err == userion.ErrInvalidPassword {
    // Password is incorrect
} else if err == userion.ErrUserNotFound {
    // User not found
} else {
    // Other error
}
```

### List Users with Filtering and Pagination

```go
// Get 10 users, skipping the first 20
users, err := userManager.ListUsers(10, 20, nil, "", false)

// Filter by status
users, err := userManager.ListUsers(10, 0, map[string]interface{}{
    "status": userion.UserStatusActive,
}, "", false)

// Order by creation date, newest first
users, err := userManager.ListUsers(10, 0, nil, "created_at", true)
```

### User Status Management

```go
// Set status
err := userManager.SetUserStatusByID("user-uuid-here", userion.UserStatusSuspended)

// Enable/disable
err := userManager.EnableUserByID("user-uuid-here")
err := userManager.DisableUserByID("user-uuid-here")
```

### Delete a User

```go
err := userManager.DeleteUserByID("user-uuid-here")
// or
err := userManager.DeleteUserByUsername("johndoe")
// or
err := userManager.DeleteUserByEmail("john@example.com")
```

## Testing

The package includes comprehensive tests. To run them:

```bash
go test -v ./...
```

## Extending with Custom Database Implementations

Userion uses a clean interface design that allows you to implement your own data store by implementing the `UserManager` interface.

```go
type UserManager interface {
    // Interface methods here
}
```

## License

[Apache License 2.0](LICENSE)
