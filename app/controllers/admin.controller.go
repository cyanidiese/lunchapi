package controllers

import (
	"github.com/revel/revel"
	"net/http"
	"lunchapi/app/errors"
	"lunchapi/app/models"
	"github.com/uniplaces/carbon"
	"strings"
	"lunchapi/app/responses"
)

type AdminController struct {
	*revel.Controller
}

// @Summary Full Orders History
// @Description Get Full Orders History
// @Accept  json
// @Produce  json
// @Param from_date path string true "Start Date"
// @Param to_date path string true "End Date"
// @Success 200 {array} models.Order
// @Success 400 {object} errors.RequestError
// @Success 401 {object} errors.RequestError
// @Router /admin/history/{from_date}/to/{to_date} [get]
// @Security Authorization
// @Tags Admin
func (c AdminController) History() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	currentUser := AuthGetCurrentUser(c.Request)
	if currentUser.Role.Name != "admin" {
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(errors.ErrorForbidden("Only admin can view full orders history"))
	}

	fromDate := c.Params.Route.Get("from_date")
	toDate := c.Params.Route.Get("to_date")

	fromDateParsed, fromDateErr := carbon.Parse(carbon.DateFormat, fromDate, currentUser.Timezone)
	toDateParsed, toDateErr := carbon.Parse(carbon.DateFormat, toDate, currentUser.Timezone)

	if (fromDateErr != nil) || (toDateErr != nil) {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("Wrong date range", nil))
	}

	orders := []models.Order{}
	DB.
		Where("created_at > ? ", fromDateParsed.StartOfDay().DateTimeString()).
		Where("updated_at < ? ", toDateParsed.EndOfDay().DateTimeString()).
		Preload("User").
		Preload("MenuItem").
		Preload("MenuItem.Dish").
		Preload("MenuItem.Dish.Name").
		Preload("MenuItem.Dish.Description").
		Preload("MenuItem.Dish.Images").
		Find(&orders)

	return c.RenderJSON(orders)
}

// @Summary Disable User
// @Description Disable User By Id
// @Accept  json
// @Produce  json
// @Param id path int true "User Id"
// @Success 200 {object} responses.GeneralResponse
// @Success 401 {object} errors.RequestError
// @Router /admin/users/{id}/disable [put]
// @Security Authorization
// @Tags Admin
func (c AdminController) Disable() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	currentUser := AuthGetCurrentUser(c.Request)
	if currentUser.Role.Name != "admin" {
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(errors.ErrorForbidden("Only admin can disable users"))
	}

	var user models.User

	id := c.Params.Route.Get("id")

	DB.Where("id = ?", id).First(&user)

	if user.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find user based on your request"))
	}

	if user.Id == currentUser.Id {
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(errors.ErrorForbidden("Unable to disable yourself"))
	}

	DB.Table("users").Where("id = ?", user.Id).Updates(map[string]interface{}{"isBlocked": 1})

	return c.RenderJSON(responses.SuccessfulResponse("User has been successfully disabled"))
}

func (c AdminController) getAvailableCountsOfMenuItems(menuItems []models.MenuItem) (availableCounts map[int64]int64) {

	availableCounts = make(map[int64]int64)

	for index := range menuItems {
		itemId := menuItems[index].Id
		availableCount := menuItems[index].AvailableCount
		availableCounts[itemId] = availableCount
	}

	return
}

func (c AdminController) checkProviderIsAvailable(providerId string) (provider models.User, resultError errors.RequestError) {

	if len(providerId) == 0 {
		c.Response.Status = http.StatusNotFound
		return provider, errors.ErrorNotFound("Unable to find provider based on your request")
	}

	DB.Where("id = ?", providerId).Find(&provider)

	if provider.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return provider, errors.ErrorNotFound("Unable to find provider based on your request")
	}

	return provider, resultError
}

func (c AdminController) checkMenuIsAvailable(menuDate string, provider models.User) (menu models.Menu, resultError errors.RequestError) {

	if len(menuDate) == 0 {
		c.Response.Status = http.StatusNotFound
		return menu, errors.ErrorNotFound("Unable to find menu based on your request")
	}

	dateParsed, dateErr := carbon.Parse(carbon.DateFormat, menuDate, provider.Timezone)

	if dateErr != nil {
		c.Response.Status = http.StatusBadRequest
		return menu, errors.ErrorBadRequest("Unable to find menu based on this date", nil)
	}

	DB.
		Where("provider_id = ?", provider.Id).
		Where("date = ?", dateParsed.DateString()).
		Preload("Items").
		First(&menu)

	if menu.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return menu, errors.ErrorNotFound("Unable to find menu based on this date")
	}

	return menu, resultError
}

func (c AdminController) checkMenuIsBeforeDeadline(menuDate string, provider models.User) (menu models.Menu, resultError errors.RequestError) {

	menu, err := AdminController.checkMenuIsAvailable(c, menuDate, provider)
	if err.Status != 0 {
		return menu, err
	}

	deadline := strings.Replace(menu.DeadlineAt, "T", " ", -1)
	deadline = strings.Replace(deadline, "Z", "", -1)
	deadlineParser, dateErr := carbon.Parse(carbon.DefaultFormat, deadline, provider.Timezone)

	if dateErr != nil {
		c.Response.Status = http.StatusBadRequest
		return menu, errors.ErrorBadRequest("Unable to detect deadline of menu", nil)
	}

	if deadlineParser.IsPast() {
		c.Response.Status = http.StatusBadRequest
		return menu, errors.ErrorBadRequest("Deadline of this menu is already in past", nil)
	}

	return menu, resultError
}
