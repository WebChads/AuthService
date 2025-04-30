package entities

import "github.com/google/uuid"

type User struct {
	Id          uuid.UUID
	PhoneNumber string
	UserRole    string
}
