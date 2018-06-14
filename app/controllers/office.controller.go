package controllers

import (
	"github.com/revel/revel"
	"lunchapi/app/models"
	"github.com/revel/revel/cache"
	"time"
	"net/http"
	"lunchapi/app/errors"
)

type OfficeController struct {
	*revel.Controller
}

// @Summary Get Offices
// @Description Get Offices List
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Office
// @Success 401 {object} errors.RequestError
// @Router /offices/index [get]
// @Tags Offices
func (c OfficeController) Index() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	offices := []models.Office{}

	if err := cache.Get("office_index", &offices); err != nil {

		DB.Where("is_provider != ?", 1).Preload("Title").Find(&offices)
		
		go cache.Set("office_index", offices, 30*time.Minute)
	}

	return c.RenderJSON(offices)
}