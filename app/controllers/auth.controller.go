package controllers

import (
	"github.com/revel/revel"
	"lunchapi/app/responses"
	"lunchapi/app/models"
	"net/http"
    "crypto/rand"
	"strings"
	"github.com/revel/revel/cache"
	"time"
	"fmt"
)

type AuthController struct {
	*revel.Controller
}

// @Summary Login
// @Description Get Login Info
// @Accept  json
// @Produce  json
// @Param body body requests.AuthLoginRequest true "Request Body"
// @Success 200 {array} responses.AuthLoginResponse
// @Success 401 {array} responses.AuthLoginResponse
// @Router /auth/login [post]
// @Tags Auth
func (c AuthController) Login() revel.Result {

	response := responses.AuthLoginResponse{}
	user := models.User{}

	var jsonData map[string]string
	c.Params.BindJSON(&jsonData)
	//return c.RenderJSON(jsonData)

	//TODO : detect real login credentials
	var username string

	if len(jsonData["username"]) > 0 {
		username = jsonData["username"]
	} else {
		username = c.Params.Form.Get("username")
		//password = c.Params.Form.Get("password")
	}

	DB.
		Where("`email` = ?", username).
		First(&user)

	if user.Id == 0 {
		response.Success = false
		c.Response.Status = http.StatusUnauthorized
		response.Data.Error.Message = "Sorry, your login was invalid! Please check your password and try again."
		response.Data.Error.ErrorDescription = "Invalid username and password combination"
		response.Data.Error.Error = "invalid_grant"
	} else {
		response.Success = true
		response.Data.Token = user.Token
	}

	return c.RenderJSON(response)
}

func AuthGetToken(request *revel.Request) string {
	authHeader := request.GetHttpHeader("Authorization")
	authToken := strings.Replace(authHeader, "Bearer ", "", -1)
	authToken = "998d29a66db2a80dcae77fefa0a4e503" //TODO : detect real token
	return authToken
}

func AuthGetCurrentUser(request *revel.Request) models.User {

	authToken := AuthGetToken(request)

	user := models.User{}

	cacheKey := "user_" + authToken
	if err := cache.Get(cacheKey, &user); err != nil {
		DB.
			Where("`token` = ?", authToken).
			Preload("Image").
			Preload("Role").
			First(&user)

		go cache.Set(cacheKey, user, 30*time.Minute)
	}
	return user
}

func AuthCheck(request *revel.Request) bool {
	authToken := AuthGetToken(request)
	if len(authToken) == 0 {
		return false
	}
	user := AuthGetCurrentUser(request)
	if user.Id == 0 {
		return false
	}
	return true
}

func AuthRandToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
