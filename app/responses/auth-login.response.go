package responses

type AuthLoginResponse struct {
	Success    bool   `json:"success"`
	StatusCode int64  `json:"statusCode"`
	Token      string `json:"token"`
	Message    string `json:"message"`
}
