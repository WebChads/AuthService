package dtos

type SendSmsCodeRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type VerifySmsCodeRequest struct {
	PhoneNumber string `json:"phone_number"`
	SmsCode     string `json:"sms_code"`
}
