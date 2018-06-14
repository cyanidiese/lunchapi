package controllers

import (
	"github.com/revel/revel"
	"lunchapi/app/models"
	"github.com/revel/revel/cache"
	"time"
	"net/http"
	"lunchapi/app/errors"
)

type CategoryController struct {
	*revel.Controller
}

// @Summary Get Categories
// @Description Get Categories List
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Category
// @Success 401 {object} errors.RequestError
// @Router /categories/index [get]
// @Tags Categories
func (c CategoryController) Index() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	categories := []models.Category{}

	if err := cache.Get("category_index", &categories); err != nil {

		DB.Preload("Title").Find(&categories)
		
		go cache.Set("category_index", categories, 30*time.Minute)
	}

	return c.RenderJSON(categories)
}