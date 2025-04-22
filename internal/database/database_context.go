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

var DatabaseContextObject *DatabaseContext

func InitDatabase(servicesScope *services.ServicesScope) (*DatabaseContext, error) {
	dbSettings := servicesScope.Configuration.DbSettings
	connectionString := fmt.Sprintf("Host=%s; User=%s; Password=%s", dbSettings.Host, dbSettings.User, dbSettings.Password)

	connection, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	DatabaseContextObject = &DatabaseContext{Services: servicesScope, Connection: connection}

	createDbCommand := `BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'your_database_name') THEN
        CREATE DATABASE your_database_name;
    END IF;
END`

	createDbCommand = strings.ReplaceAll(createDbCommand, "your_database_name", dbSettings.DbName)
	_, err = connection.Exec(createDbCommand)
	if err != nil {
		return nil, err
	}

	err = DatabaseContextObject.MigrateTables()
	if err != nil {
		return nil, err
	}

	return DatabaseContextObject, nil
}

func (databaseContext *DatabaseContext) MigrateTables() error {
	if databaseContext.Connection == nil {
		return errors.New("There are no connection while migration of tables")
	}

	usersTable := `CREATE TABLE users
    (
        id uuid PRIMARY KEY
        phone_number varchar(12)
        user_role varchar(25)
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
