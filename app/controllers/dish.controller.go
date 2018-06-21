package controllers

import (
	"github.com/revel/revel"
	"lunchapi/app/models"
	"net/http"
	"lunchapi/app/errors"
)

type DishController struct {
	*revel.Controller
}

// @Summary Get Dishes
// @Description Get Dishes List
// @Accept  json
// @Produce  json
// @Param provider_id query int false "Provider ID"
// @Param category_id query int false "Category ID"
// @Success 200 {array} models.Dish
// @Success 401 {object} errors.RequestError
// @Router /dishes/index [get]
// @Security Authorization
// @Tags Dishes
func (c DishController) Index() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	dishes := []models.Dish{}

	providerId := c.Params.Query.Get("provider_id")
	categoryId := c.Params.Query.Get("category_id")

	query := DB.Preload("Name").Preload("Description").Preload("Images")

	if len(providerId) > 0 {
		query = query.Where("provider_id = ?", providerId)
	}
	if len(categoryId) > 0 {
		query = query.Where("category_id = ?", categoryId)
	}

	query.Find(&dishes)

	return c.RenderJSON(dishes)
}