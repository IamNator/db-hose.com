package schema

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordData struct {
	Email           string `json:"email"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
