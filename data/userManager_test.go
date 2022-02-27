package data

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddUser(t *testing.T) {
	manager := getUsersManager()
	assert.NotNil(t, manager)

	manager.clean()

	newUser := &User{
		Name:     "UserName",
		LastName: "UserLastName",
	}

	userCount := manager.getCount()
	assert.Equal(t, 0, userCount)

	userId, err := manager.add(newUser)
	assert.Nil(t, err)
	assert.NotEmpty(t, userId)

	userCount = manager.getCount()
	assert.Equal(t, 1, userCount)
}

func TestGetUser(t *testing.T) {
	manager := getUsersManager()
	assert.NotNil(t, manager)

	manager.clean()

	newUser := &User{
		Name:     "UserName",
		LastName: "UserLastName",
	}

	userCount := manager.getCount()
	assert.Equal(t, 0, userCount)

	userId, err := manager.add(newUser)
	assert.Nil(t, err)
	assert.NotEmpty(t, userId)

	userCount = manager.getCount()
	assert.Equal(t, 1, userCount)

	user, err := manager.get(userId)
	assert.Nil(t, err)
	assert.Equal(t, "UserName", user.Name)
	assert.Equal(t, "UserLastName", user.LastName)
}
