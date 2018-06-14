package responses

type authLoginError struct {
	Message          string `json:"message"`
	ErrorDescription string `json:"error_description"`
	Error            string `json:"error"`
}

type authLoginData struct {
	Token        string         `json:"token"`
	ExpiresIn    uint64         `json:"expires_in"`
	RefreshToken string         `json:"refresh_token"`
	Error        authLoginError `json:"error"`
}

type AuthLoginResponse struct {
	Success    bool          `json:"success"`
	StatusCode int64         `json:"statusCode"`
	Data       authLoginData `json:"data"`
}
