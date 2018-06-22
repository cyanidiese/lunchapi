package controllers

import (
	"github.com/revel/revel"
	"lunchapi/app/models"
	"github.com/revel/revel/cache"
	"time"
	"net/http"
	"lunchapi/app/errors"
	"github.com/uniplaces/carbon"
	"lunchapi/app/responses"
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
// @Security Authorization
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

// @Summary Add or update Categories
// @Description Add or update Categories
// @Accept  json
// @Produce  json
// @Param body body models.Category true "Category Details"
// @Success 200 {object} models.Category
// @Success 401 {object} errors.RequestError
// @Success 403 {object} errors.RequestError
// @Router /categories/save [post]
// @Security Authorization
// @Tags Categories
func (c CategoryController) Save() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)

	if user.Role.Name != "admin" {
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(errors.ErrorForbidden("Only admin can add or update categories"))
	}

	var categoryData, category models.Category
	c.Params.BindJSON(&categoryData)

	if categoryData.Id != 0 {
		DB.
			Where("id = ?", categoryData.Id).
			Preload("Title").
			First(&category)

		if category.Id == 0 {
			c.Response.Status = http.StatusNotFound
			return c.RenderJSON(errors.ErrorNotFound("Unable to find category based on your request"))
		}

		category.Title.En = categoryData.Title.En
		category.Title.Ua = categoryData.Title.Ua
		category.Title.Ru = categoryData.Title.Ru
		category.Title.UpdatedAt = carbon.Now().Time
		categoryData.Title = category.Title
	}

	DB.Create(&categoryData)
	DB.Save(&categoryData)

	categories := []models.Category{}
	DB.Preload("Title").Find(&categories)
	go cache.Set("category_index", categories, 30*time.Minute)

	return c.RenderJSON(categoryData)
}

// @Summary Delete Category
// @Description Delete Category By Id
// @Accept  json
// @Produce  json
// @Param id path int true "Category Id"
// @Success 200 {object} responses.GeneralResponse
// @Success 401 {object} errors.RequestError
// @Router /categories/{id}/delete [delete]
// @Security Authorization
// @Tags Categories
func (c CategoryController) Delete() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)
	if user.Role.Name != "admin" {
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(errors.ErrorForbidden("Only admin can remove categories"))
	}

	var category, categoryToReplace models.Category

	id := c.Params.Route.Get("id")

	DB.Where("id = ?", id).First(&category)

	if category.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find category based on your request"))
	}

	DB.Where("id != ?", id).First(&categoryToReplace)

	if categoryToReplace.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to remove last category"))
	}

	DB.Table("dishes").Where("category_id = ?", category.Id).Updates(map[string]interface{}{"name": categoryToReplace.Id})
	DB.Where("id = ?", id).Delete(models.Category{})

	return c.RenderJSON(responses.SuccessfulResponse("Category has been successfully removed"))
}