package data

import (
	"19PMI/19PMI/logs"
	"errors"
	"gitlab.com/rwxrob/uniq"
	"sync"
)

type usersManager struct {
	logger *logs.Logger
	users  map[string]*User

	// Signals
	initTaskCh     chan *addTask
	addTaskCh      chan *addTask
	getTaskCh      chan *getTask
	updateTaskCh   chan *updateTask
	removeTaskCh   chan *removeTask
	getCountTaskCh chan *getCountTask
	cleanTaskCh    chan *cleanTask
}

type addTask struct {
	user   *User
	result chan *addTaskResponse
}

type addTaskResponse struct {
	userId string
	err    error
}

type getTask struct {
	userId string
	result chan *getTaskResponse
}

type getTaskResponse struct {
	user *User
	err  error
}

type updateTask struct {
	user *User
	ok   chan bool
}

type removeTask struct {
	userId string
	ok     chan bool
}

type getCountTask struct {
	count chan int
}

type cleanTask struct {
	done chan interface{}
}

func (m *usersManager) initClient(user *User) {
	responseCh := make(chan *addTaskResponse, 1)
	task := &addTask{
		user:   user,
		result: responseCh,
	}
	m.initTaskCh <- task

	<-responseCh

	return
}

func (m *usersManager) add(user *User) (userId string, err error) {
	responseCh := make(chan *addTaskResponse, 1)
	task := &addTask{
		user:   user,
		result: responseCh,
	}
	m.addTaskCh <- task

	response := <-responseCh
	userId = response.userId
	err = response.err

	return
}

func (m *usersManager) get(userId string) (user *User, err error) {
	responseCh := make(chan *getTaskResponse, 1)
	defer close(responseCh)

	task := &getTask{
		userId: userId,
		result: responseCh,
	}
	m.getTaskCh <- task

	response := <-responseCh
	user = response.user
	err = response.err

	return
}

func (m *usersManager) update(user *User) (ok bool) {
	responseCh := make(chan bool, 1)
	defer close(responseCh)

	task := &updateTask{
		user: user,
		ok:   responseCh,
	}
	m.updateTaskCh <- task

	ok = <-responseCh

	return
}

func (m *usersManager) remove(userId string) (ok bool) {
	responseCh := make(chan bool, 1)
	defer close(responseCh)

	task := &removeTask{
		userId: userId,
		ok:     responseCh,
	}
	m.removeTaskCh <- task

	ok = <-responseCh

	return
}

func (m *usersManager) getCount() (count int) {
	responseCh := make(chan int, 1)
	defer close(responseCh)

	task := &getCountTask{
		count: responseCh,
	}
	m.getCountTaskCh <- task

	count = <-responseCh

	return
}

func (m *usersManager) clean() {
	responseCh := make(chan interface{}, 1)
	defer close(responseCh)

	task := &cleanTask{
		done: responseCh,
	}
	m.cleanTaskCh <- task

	<-responseCh

	return
}

/**
 * Setup singleton
 */
var uMutex = &sync.Mutex{}
var uInstance *usersManager

func getUsersManager() *usersManager {
	if uInstance == nil {
		uMutex.Lock()
		if uInstance == nil {
			uInstance = &usersManager{}
			uInstance.init()
		}
		uMutex.Unlock()
	}

	return uInstance
}

func (m *usersManager) init() {
	usersManagerLogger := logs.GetLogger().SetSource("UsersManager")
	m.logger = &usersManagerLogger

	// Pool
	m.users = make(map[string]*User, 65536)
	// Signals
	m.initTaskCh = make(chan *addTask, 128)
	m.addTaskCh = make(chan *addTask, 128)
	m.getTaskCh = make(chan *getTask, 128)
	m.updateTaskCh = make(chan *updateTask, 128)
	m.removeTaskCh = make(chan *removeTask, 128)
	m.getCountTaskCh = make(chan *getCountTask, 128)
	m.cleanTaskCh = make(chan *cleanTask, 128)

	// Start pool
	go m.run()
	go m.initClients()
}

func (m *usersManager) initClients() {
	initCh := getDataProvider().getClientsInitChan()
	for client := range initCh {
		m.initClient(client)
	}
}

/* *******************
 * ThreadSafe Map pool
 * *******************/
func (m *usersManager) run() {
	m.logger.Info().Msg(
		"started",
	)

	defer func() {
		m.logger.Info().Msg(
			"stopped",
		)
	}()

	for {
		select {
		case task := <-m.initTaskCh:
			m.users[task.user.Id] = task.user
			task.result <- &addTaskResponse{
				userId: task.user.Id,
				err:    nil,
			}

			m.logger.Info().Msgf(
				"user added from dataBase. UserId:%s PoolSize:%d",
				task.user.Id,
				len(m.users),
			)

		case task := <-m.addTaskCh:
			task.user.Id = uniq.UUID()

			m.users[task.user.Id] = task.user
			task.result <- &addTaskResponse{
				userId: task.user.Id,
				err:    nil,
			}

			dataProvider := getDataProvider()
			dataProvider.addUser(task.user)

			m.logger.Info().Msgf(
				"user added. UserId:%s PoolSize:%d",
				task.user.Id,
				len(m.users),
			)

		case task := <-m.getTaskCh:
			user, exist := m.users[task.userId]
			if !exist {
				err := errors.New("user not found")
				task.result <- &getTaskResponse{
					user: nil,
					err:  err,
				}

				m.logger.Error().Msgf(
					"user not found. UserId:%s",
					task.userId,
				)

				continue
			}

			task.result <- &getTaskResponse{
				user: user,
				err:  nil,
			}

			m.logger.Info().Msgf(
				"user returned. UserId:%s",
				task.userId,
			)

		case task := <-m.updateTaskCh:
			userFromMap, exist := m.users[task.user.Id]
			if !exist {
				task.ok <- false

				continue
			}

			userFromMap.Name = task.user.Name
			userFromMap.LastName = task.user.LastName

			task.ok <- true

			dataProvider := getDataProvider()
			dataProvider.updateUser(task.user)

			m.logger.Info().Msgf(
				"user updated. UserId:%s",
				task.user.Id,
			)

		case task := <-m.removeTaskCh:
			_, exist := m.users[task.userId]
			if !exist {
				task.ok <- false

				continue
			}

			delete(m.users, task.userId)
			task.ok <- true

			dataProvider := getDataProvider()
			dataProvider.removeUser(task.userId)

			m.logger.Info().Msgf(
				"user removed. UserId:%s PoolSize:%d",
				task.userId,
				len(m.users),
			)

		case task := <-m.getCountTaskCh:
			task.count <- len(m.users)

		case task := <-m.cleanTaskCh:
			m.users = make(map[string]*User)
			task.done <- struct{}{}
		}
	}
}
