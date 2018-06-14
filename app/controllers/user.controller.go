package controllers

import (
	"github.com/revel/revel"
	"net/http"
	"lunchapi/app/errors"
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
