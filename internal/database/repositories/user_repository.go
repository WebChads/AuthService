package repositories

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/WebChads/AuthService/internal/models/entities"
)

type UserRepository interface {
	AddNewUser(user *entities.User) error

	// If user does not exists - returns nil, nil
	GetUserByPhoneNumber(phoneNumber string) (*entities.User, error)
}

// Implementation of UserRepository for database/sql + PostgreSQL
type PgUserRepository struct {
	connection *sql.DB
}

func (repository *PgUserRepository) AddNewUser(user *entities.User) error {
	amountOfUsersWithThisPhoneNumber, err := repository.getCountOfUserWithThatPhoneNumber(user.PhoneNumber)
	if err != nil {
		return fmt.Errorf("while adding new user happened error: %w", err)
	}

	if amountOfUsersWithThisPhoneNumber != 0 {
		return errors.New("while adding new user happened error: there are already user with that phone number")
	}

	addUserQuery := "INSERT INTO users VALUES ($1, $2, $3)"
	_, err = repository.connection.Exec(addUserQuery, user.Id, user.PhoneNumber, user.UserRole)

	return err
}

func (repository *PgUserRepository) GetUserByPhoneNumber(phoneNumber string) (*entities.User, error) {
	countUsers, err := repository.getCountOfUserWithThatPhoneNumber(phoneNumber)
	if err != nil {
		return nil, err
	}

	if countUsers == 0 {
		return nil, nil
	} else if countUsers > 1 {
		return nil, fmt.Errorf("while retrieving user with phone number %s happened error: there are more than one user with that phone number", phoneNumber)
	}

	user := &entities.User{}
	userQuery := "SELECT id, phone_number, user_role FROM users WHERE phone_number = $1"
	err = repository.connection.QueryRow(userQuery, phoneNumber).Scan(&user.Id, &user.PhoneNumber, &user.UserRole)
	if err != nil {
		return nil, fmt.Errorf("while retrieving user with phone number %s happened error: %w", phoneNumber, err)
	}

	return user, nil
}

func (repository *PgUserRepository) getCountOfUserWithThatPhoneNumber(phoneNumber string) (int, error) {
	countQuery := "SELECT COUNT(*) FROM users WHERE phone_number = $1"

	var amountOfUsersWithThisPhoneNumber int
	err := repository.connection.QueryRow(countQuery, phoneNumber).Scan(&amountOfUsersWithThisPhoneNumber)
	if err != nil {
		return 0, fmt.Errorf("while counting amount of users with phone number %s happened error: %w", phoneNumber, err)
	}

	return amountOfUsersWithThisPhoneNumber, nil
}
