package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/WebChads/AuthService/internal/services"
	_ "github.com/lib/pq"
)

type DatabaseContext struct {
	Services   *services.ServicesScope
	Connection *sql.DB
}

func InitDatabase(servicesScope *services.ServicesScope) (*DatabaseContext, error) {
	dbSettings := servicesScope.Configuration.DbSettings
	connectionString := fmt.Sprintf("postgres://%s:%s@%s/postgres?sslmode=disable", dbSettings.User, dbSettings.Password, dbSettings.Host)

	connection, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	doesDbExists, err := checkIfDbExists(connection, dbSettings.DbName)
	if err != nil {
		return nil, err
	}

	if !doesDbExists {
		err = createDatabase(connection, dbSettings.DbName)
		if err != nil {
			return nil, err
		}
	}

	connectionString = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbSettings.User, dbSettings.Password, dbSettings.Host, dbSettings.DbName)
	connection, err = sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	databaseContextObject := &DatabaseContext{Services: servicesScope, Connection: connection}

	err = databaseContextObject.migrateTables()
	if err != nil {
		return nil, err
	}

	return databaseContextObject, nil
}

func checkIfDbExists(connection *sql.DB, dbName string) (bool, error) {
	checkifDbExistsQuery := "SELECT true FROM pg_database WHERE datname = $1"

	res, err := connection.Query(checkifDbExistsQuery, dbName)
	if err != nil {
		return false, err
	}

	var exists bool
	res.Scan(&exists)

	return exists, nil
}

func createDatabase(connection *sql.DB, dbName string) error {
	createDbCommand := "CREATE DATABASE $1"
	_, err := connection.Exec(createDbCommand, dbName)
	if err != nil {
		return err
	}
	return nil
}

func (databaseContext *DatabaseContext) migrateTables() error {
	if databaseContext.Connection == nil {
		return errors.New("there are no connection while migration of tables")
	}

	isUsersExists, err := databaseContext.checkIfTableExists("users")
	if err != nil {
		return err
	}

	if !isUsersExists {
		err = databaseContext.createTableUsers()
		if err != nil {
			return err
		}
	}

	err = databaseContext.createIndexOnTableUsers()
	if err != nil {
		return err
	}

	return nil
}

func (databaseContext *DatabaseContext) checkIfTableExists(tableName string) (bool, error) {
	sqlQuery := `SELECT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = $1
    )`

	var exists bool
	err := databaseContext.Connection.QueryRow(sqlQuery, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if table exists: %w", err)
	}

	return exists, nil
}

func (databaseContext *DatabaseContext) createTableUsers() error {
	usersTable := `CREATE TABLE users
    (
        id uuid PRIMARY KEY NOT NULL,
        phone_number varchar(12) NOT NULL,
        user_role varchar(25) NOT NULL
    )
`
	_, err := databaseContext.Connection.Exec(usersTable)
	if err != nil {
		return err
	}

	return nil
}

func (databaseContext *DatabaseContext) createIndexOnTableUsers() error {
	usersIndex := "CREATE INDEX index_users_phone_number ON users (phone_number)"
	_, err := databaseContext.Connection.Exec(usersIndex)
	if err != nil {
		return err
	}

	return nil
}
