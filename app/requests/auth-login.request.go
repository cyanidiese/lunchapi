package requests

type AuthLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
