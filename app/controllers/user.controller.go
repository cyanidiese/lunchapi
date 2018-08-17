package controllers

import (
	"github.com/revel/revel"
	"lunchapi/app/errors"
	"lunchapi/app/models"
	"net/http"
	"strings"
	"github.com/revel/revel/cache"
	"time"
)

type UserController struct {
	*revel.Controller
}

// @Summary User Info
// @Description Get Login Info
// @Accept  json
// @Produce  json
// @Success 200 {object} models.User
// @Success 401 {object} errors.RequestError
// @Router /user/index [get]
// @Security Authorization
// @Tags User
func (c UserController) Index() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)

	return c.RenderJSON(user)
}

// @Summary Update Language
// @Description Update preferred language of user
// @Accept  json
// @Produce  json
// @Param lang path string true "Language (EN, UA, RU)"
// @Success 200 {object} models.User
// @Success 401 {object} errors.RequestError
// @Router /user/language/{lang} [post]
// @Security Authorization
// @Tags User
func (c UserController) Language() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)

	lang := c.Params.Route.Get("lang")
	lang = strings.ToLower(lang)

	if lang != "en" && lang != "ua" && lang != "ru" {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("You cannot assign wrong language to profile", nil))
	}

	user.Language = lang
	DB.Save(&user)

	go cache.Set("user_" + user.Token, user, 5*time.Minute)

	return c.RenderJSON(user)
}

// @Summary Update Provider
// @Description Update preferred provider of user
// @Accept  json
// @Produce  json
// @Param provider_id path integer true "Provider ID to set"
// @Success 200 {object} models.User
// @Success 401 {object} errors.RequestError
// @Router /user/provider/{provider_id} [post]
// @Security Authorization
// @Tags User
func (c UserController) Provider() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)
	provider := models.User{}

	providerId := c.Params.Route.Get("provider_id")

	if len(providerId) == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find provider based on your request"))
	}

	DB.Where("id = ?", providerId).Find(&provider)

	if provider.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find provider based on your request"))
	}

	user.ProviderId = provider.Id
	DB.Save(&user)

	go cache.Set("user_" + user.Token, user, 5*time.Minute)

	return c.RenderJSON(user)
}
