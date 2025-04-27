package dtos

type RegisterRequest struct {
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
}
