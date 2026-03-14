package gateway

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type passwordResetRequestRequest struct {
	Email string `json:"email"`
}

type passwordResetConfirmationRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"password"`
}

type createClientAccountRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth int64  `json:"date_of_birth"`
	Gender      string `json:"gender"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	Password    string `json:"password"`
}

type createEmployeeAccountRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth int64  `json:"date_of_birth"`
	Gender      string `json:"gender"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	Username    string `json:"username"`
	Position    string `json:"position"`
	Department  string `json:"department"`
	Password    string `json:"password"`
}
