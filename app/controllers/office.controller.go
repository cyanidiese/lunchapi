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
// @Security Authorization
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

// @Summary Add or update Offices
// @Description Add or update Offices
// @Accept  json
// @Produce  json
// @Param body body models.Office true "Office Details"
// @Success 200 {object} models.Office
// @Success 401 {object} errors.RequestError
// @Success 403 {object} errors.RequestError
// @Router /offices/save [post]
// @Security Authorization
// @Tags Offices
func (c OfficeController) Save() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)

	if user.Role.Name != "admin" {
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(errors.ErrorForbidden("Only admin can add or update offices"))
	}

	var officeData, office models.Office
	c.Params.BindJSON(&officeData)

	if officeData.Id != 0 {
		DB.
			Where("id = ?", officeData.Id).
			Preload("Title").
			First(&office)

		if office.Id == 0 {
			c.Response.Status = http.StatusNotFound
			return c.RenderJSON(errors.ErrorNotFound("Unable to find office based on your request"))
		}

		office.Title.En = officeData.Title.En
		office.Title.Ua = officeData.Title.Ua
		office.Title.Ru = officeData.Title.Ru
		office.Title.UpdatedAt = carbon.Now().Time
		officeData.Title = office.Title
	}

	DB.Create(&officeData)
	DB.Save(&officeData)

	offices := []models.Office{}
	DB.Preload("Title").Find(&offices)
	go cache.Set("office_index", offices, 30*time.Minute)

	return c.RenderJSON(officeData)
}

// @Summary Delete Office
// @Description Delete Office By Id
// @Accept  json
// @Produce  json
// @Param id path int true "Office Id"
// @Success 200 {object} responses.GeneralResponse
// @Success 401 {object} errors.RequestError
// @Router /offices/{id}/delete [delete]
// @Security Authorization
// @Tags Offices
func (c OfficeController) Delete() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)
	if user.Role.Name != "admin" {
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(errors.ErrorForbidden("Only admin can remove offices"))
	}

	var office, officeToReplace models.Office

	id := c.Params.Route.Get("id")

	DB.Where("id = ?", id).First(&office)

	if office.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find office based on your request"))
	}

	DB.
		Where("id != ?", id).
		Where("is_provider = ?", office.IsProvider).
		First(&officeToReplace)

	if officeToReplace.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to remove last office of this type"))
	}

	DB.Table("users").Where("office_id = ?", office.Id).Updates(map[string]interface{}{"name": officeToReplace.Id})
	DB.Where("id = ?", id).Delete(models.Office{})

	return c.RenderJSON(responses.SuccessfulResponse("Office has been successfully removed"))
}