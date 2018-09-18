package controllers

import (
	"github.com/revel/revel"
	"net/http"
	"lunchapi/app/errors"
)

func checkUser(c *revel.Controller) revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}
	return nil
}

func init() {
	revel.OnAppStart(InitDB)
	revel.InterceptFunc(checkUser, revel.BEFORE, &AdminController{})
	revel.InterceptFunc(checkUser, revel.BEFORE, &CategoryController{})
	revel.InterceptFunc(checkUser, revel.BEFORE, &CommentController{})
	revel.InterceptFunc(checkUser, revel.BEFORE, &DishController{})
	revel.InterceptFunc(checkUser, revel.BEFORE, &MasterController{})
	revel.InterceptFunc(checkUser, revel.BEFORE, &NotificationController{})
	revel.InterceptFunc(checkUser, revel.BEFORE, &OfficeController{})
	revel.InterceptFunc(checkUser, revel.BEFORE, &ProviderController{})
	revel.InterceptFunc(checkUser, revel.BEFORE, &UserController{})
}
