package postgres

import (
	"context"
	"testing"

	"github.com/cc1024201/opamp-platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_CreateUser(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "user",
		IsActive: true,
	}

	err := testStore.CreateUser(ctx, user)
	require.NoError(t, err)
	assert.NotZero(t, user.ID, "User ID should be set after creation")

	// Verify creation
	retrieved, err := testStore.GetUserByUsername(ctx, user.Username)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, user.Username, retrieved.Username)
	assert.Equal(t, user.Email, retrieved.Email)
	assert.Equal(t, user.Role, retrieved.Role)
}

func TestStore_GetUserByUsername(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		username string
		setup    func()
		wantUser bool
		wantErr  bool
	}{
		{
			name:     "existing user",
			username: "existinguser",
			setup: func() {
				user := &model.User{
					Username: "existinguser",
					Email:    "existing@example.com",
					Password: "hashed_password",
					Role:     "user",
				}
				testStore.CreateUser(ctx, user)
			},
			wantUser: true,
			wantErr:  false,
		},
		{
			name:     "non-existent user",
			username: "nonexistent",
			setup:    func() {},
			wantUser: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupDatabase(t)
			tt.setup()

			user, err := testStore.GetUserByUsername(ctx, tt.username)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantUser {
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username)
			} else {
				assert.Nil(t, user)
			}
		})
	}
}

func TestStore_GetUserByEmail(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Create test user
	user := &model.User{
		Username: "emailuser",
		Email:    "email@example.com",
		Password: "hashed_password",
		Role:     "user",
	}
	err := testStore.CreateUser(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name     string
		email    string
		wantUser bool
	}{
		{
			name:     "existing email",
			email:    "email@example.com",
			wantUser: true,
		},
		{
			name:     "non-existent email",
			email:    "nonexistent@example.com",
			wantUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := testStore.GetUserByEmail(ctx, tt.email)
			assert.NoError(t, err)

			if tt.wantUser {
				assert.NotNil(t, found)
				assert.Equal(t, tt.email, found.Email)
			} else {
				assert.Nil(t, found)
			}
		})
	}
}

func TestStore_GetUserByID(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Create test user
	user := &model.User{
		Username: "iduser",
		Email:    "id@example.com",
		Password: "hashed_password",
		Role:     "admin",
	}
	err := testStore.CreateUser(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name     string
		id       uint
		wantUser bool
	}{
		{
			name:     "existing user",
			id:       user.ID,
			wantUser: true,
		},
		{
			name:     "non-existent user",
			id:       99999,
			wantUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := testStore.GetUserByID(ctx, tt.id)
			assert.NoError(t, err)

			if tt.wantUser {
				assert.NotNil(t, found)
				assert.Equal(t, tt.id, found.ID)
			} else {
				assert.Nil(t, found)
			}
		})
	}
}

func TestStore_UpdateUser(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Create user
	user := &model.User{
		Username: "updateuser",
		Email:    "update@example.com",
		Password: "original_password",
		Role:     "user",
		IsActive: true,
	}
	err := testStore.CreateUser(ctx, user)
	require.NoError(t, err)

	// Update user
	user.Email = "updated@example.com"
	user.Role = "admin"
	user.IsActive = false

	err = testStore.UpdateUser(ctx, user)
	require.NoError(t, err)

	// Verify update
	updated, err := testStore.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, "updated@example.com", updated.Email)
	assert.Equal(t, "admin", updated.Role)
	assert.False(t, updated.IsActive)
	assert.Equal(t, "updateuser", updated.Username) // Username should not change
}

func TestStore_ListUsers(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Test empty list
	users, err := testStore.ListUsers(ctx)
	require.NoError(t, err)
	assert.Empty(t, users)

	// Create multiple users
	testUsers := []*model.User{
		{
			Username: "user1",
			Email:    "user1@example.com",
			Password: "password1",
			Role:     "user",
		},
		{
			Username: "user2",
			Email:    "user2@example.com",
			Password: "password2",
			Role:     "admin",
		},
		{
			Username: "user3",
			Email:    "user3@example.com",
			Password: "password3",
			Role:     "user",
		},
	}

	for _, u := range testUsers {
		err := testStore.CreateUser(ctx, u)
		require.NoError(t, err)
	}

	// List all users
	users, err = testStore.ListUsers(ctx)
	require.NoError(t, err)
	assert.Len(t, users, 3)

	// Verify usernames
	usernames := make(map[string]bool)
	for _, u := range users {
		usernames[u.Username] = true
	}
	assert.True(t, usernames["user1"])
	assert.True(t, usernames["user2"])
	assert.True(t, usernames["user3"])
}

func TestStore_DeleteUser(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Create user
	user := &model.User{
		Username: "deleteuser",
		Email:    "delete@example.com",
		Password: "password",
		Role:     "user",
	}
	err := testStore.CreateUser(ctx, user)
	require.NoError(t, err)

	userID := user.ID

	// Delete user
	err = testStore.DeleteUser(ctx, userID)
	require.NoError(t, err)

	// Verify deletion
	deleted, err := testStore.GetUserByID(ctx, userID)
	assert.NoError(t, err)
	assert.Nil(t, deleted)

	// Try to delete non-existent user (should not error)
	err = testStore.DeleteUser(ctx, 99999)
	assert.NoError(t, err)
}

func TestStore_UserConstraints(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Create first user
	user1 := &model.User{
		Username: "unique_user",
		Email:    "unique@example.com",
		Password: "password",
		Role:     "user",
	}
	err := testStore.CreateUser(ctx, user1)
	require.NoError(t, err)

	t.Run("duplicate username", func(t *testing.T) {
		user2 := &model.User{
			Username: "unique_user", // Same username
			Email:    "different@example.com",
			Password: "password",
			Role:     "user",
		}
		err := testStore.CreateUser(ctx, user2)
		assert.Error(t, err, "Should fail with duplicate username")
	})

	t.Run("duplicate email", func(t *testing.T) {
		user3 := &model.User{
			Username: "different_user",
			Email:    "unique@example.com", // Same email
			Password: "password",
			Role:     "user",
		}
		err := testStore.CreateUser(ctx, user3)
		assert.Error(t, err, "Should fail with duplicate email")
	})
}

func TestStore_UserDefaultValues(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	user := &model.User{
		Username: "defaultuser",
		Email:    "default@example.com",
		Password: "password",
		// Role not set, should default to "user"
		// IsActive not set, should default to true
	}

	err := testStore.CreateUser(ctx, user)
	require.NoError(t, err)

	retrieved, err := testStore.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	// Check defaults
	assert.Equal(t, "user", retrieved.Role)
	assert.True(t, retrieved.IsActive)
	assert.NotZero(t, retrieved.CreatedAt)
	assert.NotZero(t, retrieved.UpdatedAt)
}
