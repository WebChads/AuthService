package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
	checkifDbExistsQuery := fmt.Sprintf("SELECT true FROM pg_database WHERE datname = '%s'", dbName)
	checkifDbExistsQuery = strings.ReplaceAll(checkifDbExistsQuery, "your_database_name", dbName)

	res, err := connection.Query(checkifDbExistsQuery)
	if err != nil {
		return false, err
	}

	doesDbAlreadyExists := false
	res.Next()
	res.Scan(&doesDbAlreadyExists)

	return doesDbAlreadyExists, nil
}

func createDatabase(connection *sql.DB, dbName string) error {
	createDbCommand := fmt.Sprintf("CREATE DATABASE %s", dbName)
	_, err := connection.Exec(createDbCommand)
	if err != nil {
		return err
	}
	return nil
}

func (databaseContext *DatabaseContext) migrateTables() error {
	if databaseContext.Connection == nil {
		return errors.New("There are no connection while migration of tables")
	}

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

	usersIndex := "CREATE INDEX index_users_phone_number ON users (phone_number)"
	_, err = databaseContext.Connection.Exec(usersIndex)
	if err != nil {
		return err
	}

	return nil
}
