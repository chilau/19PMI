package user

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddUser(t *testing.T) {
	manager := GetUsersManager()
	assert.NotNil(t, manager)

	manager.Clean()

	newUser := &User{
		Name:     "UserName",
		LastName: "UserLastName",
	}

	userCount := manager.GetCount()
	assert.Equal(t, 0, userCount)

	userId, err := manager.Add(newUser)
	assert.Nil(t, err)
	assert.NotEmpty(t, userId)

	userCount = manager.GetCount()
	assert.Equal(t, 1, userCount)
}

func TestGetUser(t *testing.T) {
	manager := GetUsersManager()
	assert.NotNil(t, manager)

	manager.Clean()

	newUser := &User{
		Name:     "UserName",
		LastName: "UserLastName",
	}

	userCount := manager.GetCount()
	assert.Equal(t, 0, userCount)

	userId, err := manager.Add(newUser)
	assert.Nil(t, err)
	assert.NotEmpty(t, userId)

	userCount = manager.GetCount()
	assert.Equal(t, 1, userCount)

	user, err := manager.Get(userId)
	assert.Nil(t, err)
	assert.Equal(t, "UserName", user.Name)
	assert.Equal(t, "UserLastName", user.LastName)
}
