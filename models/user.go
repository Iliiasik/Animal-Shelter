package models

type User struct {
	ID                int    `json:"id"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	Email             string `json:"email"`
	Role              string `json:"role"`
	IsAdmin           bool   `json:"is_admin"`
	EmailConfirmed    bool   `json:"email_confirmed"`
	ConfirmationToken string `json:"confirmation_token"`
}
