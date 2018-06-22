package controllers

import (
	"github.com/revel/revel"
	"lunchapi/app/requests"
	"lunchapi/app/responses"
	"lunchapi/app/models"
	"net/http"
    "crypto/rand"
	"strings"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"github.com/revel/revel/cache"
	"time"
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

	user := models.User{}

	var authData requests.AuthLoginRequest
	c.Params.BindJSON(&authData)

	defaultLoginErrorMessage := "Sorry, your login was invalid! Please check your password and try again."

	c.Response.Status = http.StatusUnauthorized
	response := responses.AuthLoginResponse{
		Success : false,
		Message : defaultLoginErrorMessage,
	}

	DB.
		Where("`email` = ?", authData.Email).
		First(&user)

	if user.Id == 0 {
		response.Message = "NO EMAIL " + authData.Email
		return c.RenderJSON(response)
	}

	passwordMatch, err := AuthComparePasswords(user.Password, authData.Password)

	if err != nil {
		response.Message = err.Error()
		return c.RenderJSON(response)
	}
	if !passwordMatch {
		response.Message = "JUST MISMATCH"
		return c.RenderJSON(response)
	}

	c.Response.Status = http.StatusOK
	response.Success = true
	response.Token = user.Token
	response.Message = ""

	return c.RenderJSON(response)
}

// @Summary Register
// @Description Get Login Info
// @Accept  json
// @Produce  json
// @Param body body requests.AuthLoginRequest true "Request Body"
// @Success 200 {array} responses.AuthLoginResponse
// @Success 401 {array} responses.AuthLoginResponse
// @Router /auth/register [post]
// @Tags Auth
func (c AuthController) Register() revel.Result { //TODO: IMPLEMENT THIS

	user := models.User{}

	var authData requests.AuthLoginRequest
	c.Params.BindJSON(&authData)

	defaultLoginErrorMessage := "Sorry, your login was invalid! Please check your password and try again."

	c.Response.Status = http.StatusUnauthorized
	response := responses.AuthLoginResponse{
		Success : false,
		Message : defaultLoginErrorMessage,
	}

	DB.
		Where("`email` = ?", authData.Email).
		First(&user)

	if user.Id == 0 {
		response.Message = "NO EMAIL " + authData.Email
		return c.RenderJSON(response)
	}

	passwordMatch, err := AuthComparePasswords(user.Password, authData.Password)

	if err != nil {
		response.Message = err.Error()
		return c.RenderJSON(response)
	}
	if !passwordMatch {
		response.Message = "JUST MISMATCH"
		return c.RenderJSON(response)
	}

	c.Response.Status = http.StatusOK
	response.Success = true
	response.Token = user.Token
	response.Message = ""

	return c.RenderJSON(response)
}

func AuthGetToken(request *revel.Request) string {
	authHeader := request.GetHttpHeader("Authorization")
	authToken := strings.Replace(authHeader, "Bearer ", "", -1)
	return authToken
}

func AuthGetCurrentUser(request *revel.Request) models.User {

	authToken := AuthGetToken(request)

	user := models.User{}

	cacheKey := "user_" + authToken
	if err := cache.Get(cacheKey, &user); err != nil {
		DB.
			Where("`is_disabled` != 1").
			Where("`token` = ?", authToken).
			Preload("Image").
			Preload("Role").
			First(&user)

		go cache.Set(cacheKey, user, 5*time.Minute)
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

func AuthHashPassword(password string) (string, error) {

	pwd := []byte(password)
	hashed := ""

	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err == nil {
		hashed = string(hash)
	}

	return hashed, err
}

func AuthComparePasswords(storedPwd string, plainPwd string) (bool, error){

	byteHash := []byte(storedPwd)
	bytePwd := []byte(plainPwd)

	err := bcrypt.CompareHashAndPassword(byteHash, bytePwd)
	if err != nil {
		return false, err
	}

	return true, err
}