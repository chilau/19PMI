package user

import (
	"19PMI/19PMI/logs"
	"errors"
	"gitlab.com/rwxrob/uniq"
	"sync"
)

type UsersManager struct {
	logger *logs.Logger
	users  map[string]*User

	// Signals
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

func (m *UsersManager) Add(user *User) (userId string, err error) {
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

func (m *UsersManager) Get(userId string) (user *User, err error) {
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

func (m *UsersManager) Update(user *User) (ok bool) {
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

func (m *UsersManager) Remove(userId string) (ok bool) {
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

func (m *UsersManager) GetCount() (count int) {
	responseCh := make(chan int, 1)
	defer close(responseCh)

	task := &getCountTask{
		count: responseCh,
	}
	m.getCountTaskCh <- task

	count = <-responseCh

	return
}

func (m *UsersManager) Clean() {
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
var pMutex = &sync.Mutex{}
var mInstance *UsersManager

func GetUsersManager() *UsersManager {
	if mInstance == nil {
		pMutex.Lock()
		if mInstance == nil {
			mInstance = &UsersManager{}
			mInstance.init()
		}
		pMutex.Unlock()
	}

	return mInstance
}

func (m *UsersManager) init() {
	usersManagerLogger := logs.GetLogger().SetSource("UsersManager")
	m.logger = &usersManagerLogger

	// Pool
	m.users = make(map[string]*User, 65536)
	// Signals
	m.addTaskCh = make(chan *addTask)
	m.getTaskCh = make(chan *getTask)
	m.updateTaskCh = make(chan *updateTask)
	m.removeTaskCh = make(chan *removeTask)
	m.getCountTaskCh = make(chan *getCountTask)
	m.cleanTaskCh = make(chan *cleanTask)

	// Start pool
	go m.run()
}

/* *******************
 * ThreadSafe Map pool
 * *******************/
func (m *UsersManager) run() {
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
		case task := <-m.addTaskCh:
			task.user.Id = uniq.UUID()

			m.users[task.user.Id] = task.user
			task.result <- &addTaskResponse{
				userId: task.user.Id,
				err:    nil,
			}

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
