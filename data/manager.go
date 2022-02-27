package data

import (
	"19PMI/19PMI/logs"
	"errors"
	"sync"
)

type Manager struct {
	logger       *logs.Logger
	usersManager *usersManager
}

func (m *Manager) Init() {
	m.usersManager = getUsersManager()
}

func (m *Manager) Add(user *User) (userId string, err error) {
	if m.usersManager == nil {
		err = errors.New("user manager is nil")
		return
	}

	if user == nil {
		err = errors.New("user is nil")
		return
	}

	if user.Name == "" {
		err = errors.New("user name is empty")
		return
	}

	if user.LastName == "" {
		err = errors.New("user lastName is empty")
		return
	}

	return m.usersManager.add(user)
}

func (m *Manager) Get(userId string) (user *User, err error) {
	if m.usersManager == nil {
		err = errors.New("user manager is nil")
		return
	}

	return m.usersManager.get(userId)
}

func (m *Manager) Update(user *User) (err error) {
	if m.usersManager == nil {
		err = errors.New("user manager is nil")
		return
	}

	if user == nil {
		err = errors.New("user is nil")
		return
	}

	if user.Name == "" {
		err = errors.New("user name is empty")
		return
	}

	if user.LastName == "" {
		err = errors.New("user lastName is empty")
		return
	}

	ok := m.usersManager.update(user)
	if !ok {
		err = errors.New("user not updated")
		return
	}

	return
}

func (m *Manager) Remove(userId string) (err error) {
	if m.usersManager == nil {
		err = errors.New("user manager is nil")
		return
	}

	if userId == "" {
		err = errors.New("user id is empty")
		return
	}

	ok := m.usersManager.remove(userId)
	if !ok {
		err = errors.New("user not removed")
		return
	}

	return
}

var mOnceDo = &sync.Once{}
var mInstance *Manager

func GetManager() *Manager {
	mOnceDo.Do(
		func() {
			mInstance = new(Manager)

			loggerInstance := logs.GetLogger().SetSource("MainManager")
			mInstance.logger = &loggerInstance
		},
	)

	return mInstance
}
