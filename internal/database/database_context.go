package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/WebChads/AuthService/internal/services"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type DatabaseContext struct {
	Connection *sql.DB
}

func InitDatabase(databaseConfig *services.DatabaseConfig) (*DatabaseContext, error) {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s/postgres?sslmode=disable", databaseConfig.User, databaseConfig.Password, databaseConfig.Host)

	connection, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("conn 1: %w", err)
	}

	doesDbExists, err := checkIfDbExists(connection, databaseConfig.DbName)
	if err != nil {
		return nil, fmt.Errorf("conn doesDbExists: %w", err)
	}

	if !doesDbExists {
		err = createDatabase(connection, databaseConfig.DbName)
		if err != nil {
			return nil, fmt.Errorf("conn createDatabase: %w", err)
		}
	}

	connectionString = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", databaseConfig.User, databaseConfig.Password, databaseConfig.Host, databaseConfig.DbName)
	connection, err = sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("conn 2: %w", err)
	}

	databaseContextObject := &DatabaseContext{Connection: connection}

	err = databaseContextObject.migrateTables()
	if err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return databaseContextObject, nil
}

func checkIfDbExists(connection *sql.DB, dbName string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"

	err := connection.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query failed: %w", err)
	}

	return exists, nil
}

func createDatabase(connection *sql.DB, dbName string) error {
	createDbCommand := fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(dbName))
	_, err := connection.Exec(createDbCommand)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
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
	exists, err := databaseContext.checkIfIndexExists("index_users_phone_number")
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if exists {
		return nil
	}

	usersIndex := "CREATE INDEX index_users_phone_number ON users (phone_number)"
	_, err = databaseContext.Connection.Exec(usersIndex)
	if err != nil {
		return err
	}

	return nil
}

func (databaseContext *DatabaseContext) checkIfIndexExists(indexName string) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1 FROM pg_indexes 
            WHERE indexname = $1 
            AND tablename = 'users'
        )`

	var exists bool
	err := databaseContext.Connection.QueryRow(query, indexName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}

	return exists, nil
}
