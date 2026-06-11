package auth

type PassStatusResponse struct {
	HasPassword bool `json:"has_password"`
}

type AuthIdentityResponse struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
}
