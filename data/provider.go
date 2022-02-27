package data

import (
	"19PMI/19PMI/config"
	"19PMI/19PMI/logs"
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	clientInitCh = make(chan *User, 128)
)

type provider struct {
	logger   *logs.Logger
	dataBase *sql.DB
}

func (p *provider) init() {
	dataProviderLogger := logs.GetLogger().SetSource("DataProvider")
	p.logger = &dataProviderLogger

	var err error
	defer func() {
		if err != nil {
			p.logger.Error().Msg(err.Error())
		}
	}()

	appConfig := config.GetApplicationConfiguration()
	p.dataBase, err = sql.Open("sqlite3", appConfig.DBPath)
	if err != nil {
		return
	}

	p.getClients()
}

func (p *provider) getClientsInitChan() chan *User {
	return clientInitCh
}

func (p *provider) getClients() {
	defer func() {
		close(clientInitCh)
	}()

	sqlGetClients := `SELECT * FROM clients`
	usersStatement, err := p.dataBase.Prepare(sqlGetClients)
	if err != nil {
		return
	}

	userRows, err := usersStatement.Query()
	if err != nil {
		return
	}

	for userRows.Next() {
		var userId string
		var userName string
		var userLastName string

		_ = userRows.Scan(&userId, &userName, &userLastName)
		user := &User{
			Id:       userId,
			Name:     userName,
			LastName: userLastName,
		}

		clientInitCh <- user
	}
}

func (p *provider) addUser(user *User) {
	if p.dataBase == nil {
		return
	}

	sqlAddClient := `
	INSERT OR REPLACE INTO clients(
		id,
		name,
		lastName
	) values(?, ?, ?)
	`

	statement, err := p.dataBase.Prepare(sqlAddClient)
	if err != nil {
		return
	}

	defer func() {
		_ = statement.Close()
	}()

	_, err = statement.Exec(user.Id, user.Name, user.LastName)
	if err != nil {
		return
	}

	return
}

func (p *provider) updateUser(user *User) {
	if p.dataBase == nil {
		return
	}

	sqlUpdateClient := `UPDATE clients SET name=?, lastName=? WHERE id=?`
	statement, err := p.dataBase.Prepare(sqlUpdateClient)
	if err != nil {
		return
	}

	defer func() {
		_ = statement.Close()
	}()

	_, err = statement.Exec(user.Name, user.LastName, user.Id)
	if err != nil {
		return
	}

	return
}

func (p *provider) removeUser(userId string) {
	if p.dataBase == nil {
		return
	}

	sqlUpdateClient := `DELETE FROM clients WHERE id=?`
	statement, err := p.dataBase.Prepare(sqlUpdateClient)
	if err != nil {
		return
	}

	defer func() {
		_ = statement.Close()
	}()

	_, err = statement.Exec(userId)
	if err != nil {
		return
	}

	return
}

func (p *provider) close() {
	if p.dataBase == nil {
		return
	}

	_ = p.dataBase.Close()
}

/**
 * Setup singleton
 */
var onceDo = &sync.Once{}
var instance *provider

func getDataProvider() *provider {
	onceDo.Do(
		func() {
			instance = &provider{}
			instance.init()
		},
	)

	return instance
}
