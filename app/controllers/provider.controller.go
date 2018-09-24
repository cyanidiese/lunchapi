package controllers

import (
	"github.com/revel/revel"
	"lunchapi/app/models"
	"lunchapi/app/errors"
	"github.com/revel/revel/cache"
	"time"
	"net/http"
	"github.com/uniplaces/carbon"
	"lunchapi/app/requests"
	"strings"
	"lunchapi/app/responses"
	"strconv"
)

type ProviderController struct {
	*revel.Controller
}

// @Summary Get Providers
// @Description Get Providers List
// @Accept  json
// @Produce  json
// @Success 200 {array} models.User
// @Success 401 {object} errors.RequestError
// @Router /providers/index [get]
// @Security Authorization
// @Tags Providers
func (c ProviderController) Index() revel.Result {

	providerRole := models.Role{}
	providers := []models.User{}

	if err := cache.Get("provider_index", &providers); err != nil {

		DB.Where("name = ?", "provider").Find(&providerRole)

		DB.
			Model(providerRole).
			Preload("Role").
			Preload("Image").
			Preload("Office").
			Preload("Office.Title").
			Related(&providers)

		go cache.Set("provider_index", providers, 30*time.Minute)
	}

	return c.RenderJSON(providers)
}

// @Summary Add or update Providers
// @Description Add or update Providers
// @Accept  json
// @Produce  json
// @Param body  body requests.UpdateProfileRequest true "ProviderDetails"
// @Success 200 {object} models.User
// @Success 401 {object} errors.RequestError
// @Success 403 {object} errors.RequestError
// @Success 404 {object} errors.RequestError
// @Router /providers/save [post]
// @Security Authorization
// @Tags Providers
func (c ProviderController) Save() revel.Result {

	user := AuthGetCurrentUser(c.Request)

	var providerData requests.UpdateProfileRequest
	var provider models.User
	c.Params.BindJSON(&providerData)

	if (providerData.Id == 0) && user.Role.Name != "admin" {
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(errors.ErrorForbidden("You have no permissions to create new providers"))
	}

	creatingNewProvider := false

	if providerData.Id == 0 {
		creatingNewProvider = true

		var providerRole models.Role
		DB.Where("`name` = ?", "provider").First(&providerRole)

		provider = models.User{
			RoleId: providerRole.Id,
			Token: AuthRandToken(),
			Email: "admin@test.lunch",//TODO: change to real
		}

	} else {
		DB.
			Where("id = ?", providerData.Id).
			Where("`is_disabled` != 1").
			First(&provider)

		if providerData.Id == 0 {
			c.Response.Status = http.StatusNotFound
			return c.RenderJSON(errors.ErrorNotFound("Unable to find provider based on your request"))
		}
	}

	appliedChanges, resultError, respStatus  := UpdateUserData(providerData, &provider, creatingNewProvider)
	if respStatus != 0 {
		c.Response.Status = respStatus
		return c.RenderJSON(resultError)
	}

	if appliedChanges {

		DB.Save(&provider)

		cacheKey := "user_" + provider.Token
		DB.
			Where("`is_disabled` != 1").
			Where("`token` = ?", provider.Token).
			Preload("Image").
			Preload("Role").
			First(&provider)

		go cache.Set(cacheKey, provider, 30*time.Minute)

	}

	return c.RenderJSON(provider)
}

// @Summary Get Provider Profile
// @Description Get Provider Profile
// @Accept  json
// @Produce  json
// @Param provider_id path int true "Provider Id"
// @Success 200 {object} models.User
// @Success 401 {object} errors.RequestError
// @Router /provider/{profile_id} [get]
// @Security Authorization
// @Tags Providers
func (c ProviderController) Profile() revel.Result {

	providerId := c.Params.Route.Get("provider_id")

	provider := models.User{}

	cacheKey := "provider_" + providerId
	if err := cache.Get(cacheKey, &provider); err != nil {

		if len(providerId) == 0 {
			c.Response.Status = http.StatusNotFound
			return c.RenderJSON(errors.ErrorNotFound("Unable to find provider based on your request"))
		}

		DB.
			Where("id = ?", providerId).
			Preload("Role").
			Preload("Image").
			Preload("Office").
			Preload("Office.Title").
			Find(&provider)

		if provider.Role.Name != "provider" {
			c.Response.Status = http.StatusNotFound
			return c.RenderJSON(errors.ErrorNotFound("Unable to find provider based on your request"))
		}

		go cache.Set(cacheKey, provider, 10*time.Minute)
	}

	return c.RenderJSON(provider)
}

// @Summary Get List of Provider's Menus
// @Description Get List of Provider's Menus
// @Accept  json
// @Produce  json
// @Param provider_id path int true "Provider Id"
// @Success 200 {array} models.Menu
// @Success 401 {object} errors.RequestError
// @Router /provider/{provider_id}/menus [get]
// @Security Authorization
// @Tags Menus
func (c ProviderController) Menus() revel.Result {

	providerId := c.Params.Route.Get("provider_id")

	if len(providerId) == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find provider based on your request"))
	}

	provider := models.User{}
	menus := []models.Menu{}

	DB.Where("id = ?", providerId).Find(&provider)

	if provider.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find provider based on your request"))
	}

	DB.
		Where("provider_id = ?", providerId).
		Preload("Items").
		Preload("Items.Dish").
		Preload("Items.Dish.Name").
		Preload("Items.Dish.Images").
		Find(&menus)

	return c.RenderJSON(menus)
}

// @Summary Get Provider's Menu by Date
// @Description Get Provider's Menu by Date
// @Accept  json
// @Produce  json
// @Param provider_id path int true "Provider Id"
// @Param date path string true "Menu Date"
// @Success 200 {object} models.Menu
// @Success 401 {object} errors.RequestError
// @Router /provider/{provider_id}/menus/{date} [get]
// @Security Authorization
// @Tags Menus
func (c ProviderController) Menu() revel.Result {

	provider := models.User{}
	menu := models.Menu{}

	providerId := c.Params.Route.Get("provider_id")
	date := c.Params.Route.Get("date")

	if len(providerId) == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find provider based on your request"))
	}

	DB.Where("id = ?", providerId).Find(&provider)

	if provider.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find provider based on your request"))
	}

	query := DB.
		Where("provider_id = ?", provider.Id)

	if !provider.IsShop {

		dateParsed, dateErr := carbon.Parse(carbon.DateFormat, date, provider.Timezone)

		if dateErr != nil {
			c.Response.Status = http.StatusNotFound
			return c.RenderJSON(errors.ErrorNotFound("Unable to find menu based on this date"))
		}

		query = query.Where("date = ?", dateParsed.DateString())
	}


	query.
		Preload("Items").
		Preload("Items.Dish").
		Preload("Items.Dish.Name").
		Preload("Items.Dish.Description").
		Preload("Items.Dish.Images").
		First(&menu)

	if menu.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find menu based on this date"))
	}

	return c.RenderJSON(menu)
}

// @Summary Save Menu Changes
// @Description Save Menu Changes
// @Accept  json
// @Produce  json
// @Param provider_id path int true "Provider Id"
// @Param date path string true "Menu Date"
// @Param body body requests.MenuUpdateRequest true "Menu Items Description"
// @Success 200 {object} models.Menu
// @Success 400 {object} errors.RequestError
// @Success 401 {object} errors.RequestError
// @Success 403 {object} errors.RequestError
// @Router /provider/{provider_id}/menus/{date}/save [post]
// @Security Authorization
// @Tags Menus
func (c ProviderController) SaveMenu() revel.Result {

	if err := ProviderController.checkProviderPermissions(c); err.Status != 0 {
		return c.RenderJSON(err)
	}

	var requestBody requests.MenuUpdateRequest
	c.Params.BindJSON(&requestBody)
	date := c.Params.Route.Get("date")

	newMenu, updateError := ProviderController.updateMenuData(c, date, requestBody)
	if updateError.Status != 0 {
		return c.RenderJSON(updateError)
	}

	return c.RenderJSON(newMenu)
}

// @Summary Clone Menu To Other Date
// @Description Clone Menu To Other Date
// @Accept  json
// @Produce  json
// @Param provider_id path int true "Provider Id"
// @Param date path string true "Menu Date"
// @Param new path string true "New Date"
// @Success 200 {object} models.Menu
// @Success 401 {object} errors.RequestError
// @Router /provider/{provider_id}/menus/{date}/clone/{new} [post]
// @Security Authorization
// @Tags Menus
func (c ProviderController) CloneMenu() revel.Result {

	if err := ProviderController.checkProviderPermissions(c); err.Status != 0 {
		return c.RenderJSON(err)
	}

	provider := AuthGetCurrentUser(c.Request)
	if provider.IsShop {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("Shop provider cannot clone menus", nil))
	}

	date := c.Params.Route.Get("date")
	newDate := c.Params.Route.Get("new")

	requestBody, dateError := ProviderController.getMenuData(c, date, newDate)
	if dateError.Status != 0 {
		return c.RenderJSON(dateError)
	}

	newMenu, updateError := ProviderController.updateMenuData(c, newDate, requestBody)
	if updateError.Status != 0 {
		return c.RenderJSON(updateError)
	}

	return c.RenderJSON(newMenu)
}

// @Summary Remove Provider's Menu by Date
// @Description Remove Provider's Menu by Date
// @Accept  json
// @Produce  json
// @Param provider_id path int true "Provider Id"
// @Param date path string true "Menu Date"
// @Success 200 {object} responses.GeneralResponse
// @Success 401 {object} errors.RequestError
// @Success 404 {object} errors.RequestError
// @Router /provider/{provider_id}/menus/{date}/delete [delete]
// @Security Authorization
// @Tags Menus
func (c ProviderController) DeleteMenu() revel.Result {

	if err := ProviderController.checkProviderPermissions(c); err.Status != 0 {
		return c.RenderJSON(err)
	}

	provider := AuthGetCurrentUser(c.Request)
	menu := models.Menu{}

	date := c.Params.Route.Get("date")

	dateParsed, dateErr := carbon.Parse(carbon.DateFormat, date, provider.Timezone)

	if dateErr != nil {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find menu based on this date"))
	}
	DB.
		Where("provider_id = ?", provider.Id).
		Where("date = ?", dateParsed.DateString()).
		Preload("Items").
		First(&menu)


	if menu.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find menu based on this date"))
	}

	orderedItems := ProviderController.getOrderedDishesByMenuItems(c, menu.Items)

	//Check if we have ordered items
	if len(orderedItems) > 0 {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("Menu items has orders and cannot be removed", orderedItems))
	}

	DB.Where("menu_id = ?", menu.Id).Delete(models.MenuItem{})
	DB.Where("id = ?", menu.Id).Delete(models.Menu{})

	return c.RenderJSON(responses.SuccessfulResponse("Menu has been successfully removed"))
}

// @Summary Save Or Create Dish
// @Description Save dish changes or create a new one if no dish was found by ID in request body
// @Accept  json
// @Produce  json
// @Param provider_id path int true "Provider Id"
// @Param body body models.Dish true "Dish Details"
// @Success 200 {object} models.Dish
// @Success 400 {object} errors.RequestError
// @Success 401 {object} errors.RequestError
// @Success 403 {object} errors.RequestError
// @Router /provider/{provider_id}/dish/save [post]
// @Security Authorization
// @Tags Dishes
func (c ProviderController) SaveDish() revel.Result {

	if err := ProviderController.checkProviderPermissions(c); err.Status != 0 {
		return c.RenderJSON(err)
	}

	provider := AuthGetCurrentUser(c.Request)

	var dishData, dish models.Dish
	c.Params.BindJSON(&dishData)


	dishData.ProviderId = provider.Id

	if dishData.Id != 0 {
		DB.
			Where("provider_id = ?", provider.Id).
			Where("id = ?", dishData.Id).
			Preload("Name").
			Preload("Description").
			Preload("Images").
			First(&dish)

		if dish.Id == 0 {
			c.Response.Status = http.StatusNotFound
			return c.RenderJSON(errors.ErrorNotFound("Unable to find dish based on your request"))
		}

		dish.Name.En = dishData.Name.En
		dish.Name.Ua = dishData.Name.Ua
		dish.Name.Ru = dishData.Name.Ru
		dish.Name.UpdatedAt = carbon.Now().Time
		dishData.Name = dish.Name

		dish.Description.En = dishData.Description.En
		dish.Description.Ua = dishData.Description.Ua
		dish.Description.Ru = dishData.Description.Ru
		dish.Description.UpdatedAt = carbon.Now().Time
		dishData.Description = dish.Description

		for i, newImage := range dishData.Images {
			for _, oldImage := range dish.Images {
				if newImage.Guid == oldImage.Guid {
					dishData.Images[i].Id = oldImage.Id
					dishData.Images[i].CreatedAt = oldImage.CreatedAt
					dishData.Images[i].UpdatedAt = oldImage.UpdatedAt
				}
			}
		}
	}

	DB.Create(&dishData)
	DB.Save(&dishData)

	//If provider is a shop we need to update also related menu item and create menu if not exists
	if provider.IsShop {

		var menu models.Menu
		DB.
			Where("provider_id = ?", provider.Id).
			First(&menu)

		if menu.Id == 0 {

			timeNow, _ := carbon.NowInLocation(provider.Timezone)

			menu.Date = timeNow.DateString()
			menu.DeadlineAt = timeNow.DateTimeString()
			menu.DeliveryTime = timeNow.TimeString()
			menu.ProviderId = provider.Id

			DB.Create(&menu)
			DB.Save(&menu)
		}

		var menuItem models.MenuItem

		DB.
			Where("menu_id = ?", menu.Id).
			Where("dish_id = ?", dishData.Id).
			First(&menuItem)

		if menuItem.Id != 0 {
			menuItem.Price = dishData.Price
		} else {
			menuItem.MenuId = menu.Id
			menuItem.DishId = dishData.Id
			menuItem.Price = dishData.Price
			menuItem.InitialCount = 0
			menuItem.AvailableCount = 0

			DB.Create(&menuItem)
		}
		DB.Save(&menuItem)
	}

	return c.RenderJSON(dish)
}

// @Summary Remove Provider's Dish by Id
// @Description Remove Provider's Dish by Id
// @Accept  json
// @Produce  json
// @Param provider_id path int true "Provider Id"
// @Param id path string true "Dish Id"
// @Success 200 {array} responses.GeneralResponse
// @Success 401 {object} errors.RequestError
// @Success 404 {object} errors.RequestError
// @Router /provider/{provider_id}/dish/{id}/delete [delete]
// @Security Authorization
// @Tags Dishes
func (c ProviderController) DeleteDish() revel.Result {

	if err := ProviderController.checkProviderPermissions(c); err.Status != 0 {
		return c.RenderJSON(err)
	}

	provider := AuthGetCurrentUser(c.Request)
	dish := models.Dish{}

	dishId := c.Params.Route.Get("id")

	DB.
		Where("provider_id = ?", provider.Id).
		Where("id = ?", dishId).
		First(&dish)


	if dish.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find dish based on your request"))
	}

	dish.IsRemoved = !dish.IsRemoved
	DB.Save(&dish)

	messagePart := "removed"
	if !dish.IsRemoved {
		messagePart = "restored"
	}

	return c.RenderJSON(responses.SuccessfulResponse("Dish has been successfully " + messagePart))
}


// @Summary Provider Orders History
// @Description Get Master Orders History
// @Accept  json
// @Produce  json
// @Param provider_id path int true "Provider Id"
// @Param from_date path string true "Start Date"
// @Param to_date path string true "End Date"
// @Success 200 {array} models.Order
// @Success 400 {object} errors.RequestError
// @Success 401 {object} errors.RequestError
// @Router /provider/{provider_id}/history/{from_date}/to/{to_date} [get]
// @Security Authorization
// @Tags Providers
func (c ProviderController) History() revel.Result {

	//if err := ProviderController.checkProviderPermissions(c); err.Status != 0 {
	//	return c.RenderJSON(err)
	//}

	user := AuthGetCurrentUser(c.Request)

	providerIdStr := c.Params.Route.Get("provider_id")
	fromDate := c.Params.Route.Get("from_date")
	toDate := c.Params.Route.Get("to_date")

	fromDateParsed, fromDateErr := carbon.Parse(carbon.DateFormat, fromDate, user.Timezone)
	toDateParsed, toDateErr := carbon.Parse(carbon.DateFormat, toDate, user.Timezone)
	providerId, providerIdErr := strconv.ParseInt(providerIdStr, 10, 64)

	if providerIdErr != nil {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("Wrong provider id", nil))
	}
	if (fromDateErr != nil) || (toDateErr != nil) {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("Wrong date range", nil))
	}
	if (user.Role.Name == "provider") && (user.Id != providerId) {
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(errors.ErrorForbidden("You have no permissions to view orders history of this provider"))
	}

	orders := []models.Order{}

	query := DB.
		Where("created_at > ? ", fromDateParsed.StartOfDay().DateTimeString()).
		Where("updated_at < ? ", toDateParsed.EndOfDay().DateTimeString()).
		Where("item_id IN (?)",
		DB.
			Model(models.MenuItem{}).
			Select("id").Where("menu_id IN (?)",
			DB.
				Model(models.Menu{}).
				Select("id").Where("provider_id = ?", providerId).
				QueryExpr()).
			QueryExpr())

	if user.Role.Name == "master" {
		query = query.Where("user_id = ? ", user.Id)
	}

	query.Preload("Master").
		Preload("Master.Image").
		Preload("Master.Office").
		Preload("Master.Office.Title").
		Preload("MenuItem").
		Preload("MenuItem.Dish").
		Preload("MenuItem.Dish.Images").
		Preload("MenuItem.Dish.Name").
		Preload("MenuItem.Dish.Description").
		Find(&orders)

	return c.RenderJSON(orders)
}

func (c ProviderController) checkProviderPermissions() errors.RequestError {

	resultError := errors.RequestError{}

	providerId := c.Params.Route.Get("provider_id")

	user := AuthGetCurrentUser(c.Request)

	if len(providerId) == 0 {
		c.Response.Status = http.StatusNotFound
		return errors.ErrorNotFound("Unable to find provider based on your request")
	}

	provider := models.User{}

	DB.Where("id = ?", providerId).Find(&provider)

	if user.Role.Name != "admin" {
		if user.Id != provider.Id {
			c.Response.Status = http.StatusForbidden
			return errors.ErrorForbidden("You have no permissions to change any data of this user")
		}
	}

	return resultError
}

func (c ProviderController) getMenuData(date string, newDate string) (requests.MenuUpdateRequest, errors.RequestError) {

	provider := AuthGetCurrentUser(c.Request)

	result := requests.MenuUpdateRequest{}
	menu := models.Menu{}
	resultError := errors.RequestError{}

	dateParsed, dateErr := carbon.Parse(carbon.DateFormat, date, provider.Timezone)
	newDateParsed, newDateErr := carbon.Parse(carbon.DateFormat, newDate, provider.Timezone)

	if dateErr != nil {
		c.Response.Status = http.StatusBadRequest
		return result, errors.ErrorBadRequest("Unable to find menu based on this date", nil)
	}

	if newDateErr != nil {
		c.Response.Status = http.StatusBadRequest
		return result, errors.ErrorBadRequest("Wrong destination date format", nil)
	}

	differenceInDays := dateParsed.DiffInDays(newDateParsed, false)

	DB.
		Where("provider_id = ?", provider.Id).
		Where("date = ?", dateParsed.DateString()).
		Preload("Items").
		Preload("Items.Dish").
		First(&menu)

	if menu.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return result, errors.ErrorNotFound("Unable to find menu based on this date")
	}

	deadline := strings.Replace(menu.DeadlineAt, "T", " ", -1)
	deadline = strings.Replace(deadline, "Z", "", -1)

	deadlineParsed, errEE := carbon.Parse(carbon.DefaultFormat, deadline, provider.Timezone)
	if errEE != nil {
		c.Response.Status = http.StatusBadRequest
		return result, errors.ErrorBadRequest("Wrong Deadline date format ", nil)
	}

	result.DeliveryTime = menu.DeliveryTime
	result.Deadline = deadlineParsed.Copy().AddDays(int(differenceInDays)).DateTimeString()

	counts := []requests.ObjectCounter{}

	for index := range menu.Items {
		if !menu.Items[index].Dish.IsRemoved {
			dishId := menu.Items[index].DishId
			initialCount := menu.Items[index].InitialCount
			counts = append(counts, requests.ObjectCounter{Id: dishId, Count: initialCount})
		}
	}

	result.Items = counts

	return result, resultError
}

func (c ProviderController) updateMenuData(date string, menuItemsData requests.MenuUpdateRequest) (models.Menu, errors.RequestError) {

	provider := AuthGetCurrentUser(c.Request)

	menu := models.Menu{}
	resultError := errors.RequestError{}

	menuIsNew := false

	nowCarbon, err := carbon.NowInLocation(provider.Timezone)
	if err != nil {
		c.Response.Status = http.StatusBadRequest
		return menu, errors.ErrorBadRequest("Cannot cet current time in provider's timezone", nil)
	}
	dateCarbon := nowCarbon
	timeCarbon := nowCarbon
	deadlineCarbon := nowCarbon

	if provider.IsShop {

		DB.
			Where("provider_id = ?", provider.Id).
			First(&menu)

	} else {
		dateParsed, dateErr := carbon.Parse(carbon.DateFormat, date, provider.Timezone)
		deadlineParsed, deadlineErr := carbon.Parse(carbon.DefaultFormat, menuItemsData.Deadline, provider.Timezone)
		timeParsed, timeErr := carbon.Parse(carbon.TimeFormat, menuItemsData.DeliveryTime, provider.Timezone)
		deliveryDateTimeParsed, deliveryDateErr := carbon.Parse(carbon.DefaultFormat, date+" "+menuItemsData.DeliveryTime, provider.Timezone)
		if dateErr != nil {
			c.Response.Status = http.StatusNotFound
			return menu, errors.ErrorNotFound("Unable to find menu based on this date")
		}
		if (deadlineErr != nil) || (timeErr != nil) || (deliveryDateErr != nil) {
			c.Response.Status = http.StatusBadRequest
			return menu, errors.ErrorBadRequest("Wrong delivery time or deadline was set", nil)
		}
		if deadlineParsed.Gt(deliveryDateTimeParsed) {
			c.Response.Status = http.StatusBadRequest
			return menu, errors.ErrorBadRequest("Deadline cannot be after delivery time", nil)
		}
		if nowCarbon.Gt(deadlineParsed) {
			c.Response.Status = http.StatusBadRequest
			return menu, errors.ErrorBadRequest("Deadline cannot be in past", nil)
		}
		if nowCarbon.Gt(deliveryDateTimeParsed) {
			c.Response.Status = http.StatusBadRequest
			return menu, errors.ErrorBadRequest("Delivery time cannot be in past", nil)
		}

		dateCarbon = dateParsed
		timeCarbon = timeParsed
		deadlineCarbon = deadlineParsed

	}

	query := DB.Where("provider_id = ?", provider.Id)
	if !provider.IsShop {
		query = query.Where("date = ?", dateCarbon.DateString())
	}
	query.First(&menu)

	if menu.Id == 0 {

		timeNow, _ := carbon.NowInLocation(provider.Timezone)

		menu.ProviderId = provider.Id
		menu.Date = dateCarbon.DateString()
		menu.CreatedAt = timeNow.Time
		menu.UpdatedAt = timeNow.Time
		DB.Create(&menu)
		DB.Save(&menu)

		menuIsNew = true
	}

	menu.DeadlineAt = deadlineCarbon.DateTimeString()
	menu.DeliveryTime = timeCarbon.TimeString()
	DB.Save(&menu)

	dishesIds := []int64{}
	dishesCounts := make(map[int64]int64)

	for index := range menuItemsData.Items {
		dishId := menuItemsData.Items[index].Id
		dishCount := menuItemsData.Items[index].Count

		dishesIds = append(dishesIds, dishId)
		dishesCounts[dishId] = dishCount
	}

	var dishes []models.Dish

	DB.
		Where("provider_id = ?", provider.Id).
		Where("id in (?)", dishesIds).
		Find(&dishes)

	// We need remove or change menu items only if menu already exists
	if !menuIsNew {

		var itemsAlreadyInMenu, menuItemsToRemove []models.MenuItem

		DB.
			Where("menu_id = ?", menu.Id).
			Where("dish_id in (?)", dishesIds).
			Find(&itemsAlreadyInMenu)

		DB.
			Where("menu_id = ?", menu.Id).
			Where("dish_id not in (?)", dishesIds).
			Find(&menuItemsToRemove)

		orderedItemsForRemoving := ProviderController.getOrderedDishesByMenuItems(c, menuItemsToRemove)
		orderedItemsForUpdating := ProviderController.getOrderedDishesByMenuItems(c, itemsAlreadyInMenu)

		//Check if we have ordered items
		if (len(orderedItemsForRemoving) > 0) || len(orderedItemsForUpdating) > 0 {

			orderedItems := make(map[int64]int64)

			for k, v := range orderedItemsForRemoving {
				orderedItems[k] = v
			}

			//Unable to change items only if ordered more than new Count
			for index := range itemsAlreadyInMenu {
				dishId := itemsAlreadyInMenu[index].DishId
				orderedCount := itemsAlreadyInMenu[index].InitialCount - itemsAlreadyInMenu[index].AvailableCount

				newCount := dishesCounts[dishId]

				//We cannot set new count smaller than already ordered
				if newCount < orderedCount {
					orderedItems[dishId] = orderedCount
				}
			}

			if len(orderedItemsForRemoving) > 0 {
				c.Response.Status = http.StatusBadRequest
				return menu, errors.ErrorBadRequest("Menu items has orders and cannot be updated", orderedItemsForRemoving)
			}
		}

		//If it is all ok about new counts - go on

		if len(menuItemsToRemove) > 0 {
			//Remove items which are not in new set of dishes
			DB.Where("dish_id not in (?)", dishesIds).Delete(&models.MenuItem{})
		}

		if len(itemsAlreadyInMenu) > 0 {

			for index := range itemsAlreadyInMenu {

				menuItem := itemsAlreadyInMenu[index]

				itemId := menuItem.Id
				dishId := menuItem.DishId

				newCount := dishesCounts[dishId]

				difference := menuItem.InitialCount - newCount

				if difference != 0 {
					//Reduce available count too
					newAvailable := menuItem.AvailableCount - difference
					//Update counts
					DB.Model(&menuItem).Where("id = ?", itemId).Update("initial_count", newCount).Update("available_count", newAvailable)
				}
				//Remove dish from this map as we don't need to add it as new
				delete(dishesCounts, dishId)
			}
		}
	}

	if len(dishes) > 0 {

		timeNow, _ := carbon.NowInLocation(provider.Timezone)

		//Add dishes which are not in menu yet
		for _, dish := range dishes {

			if count, ok := dishesCounts[dish.Id]; ok {

				menuItem := models.MenuItem{}
				menuItem.DishId = dish.Id
				menuItem.Price = dish.Price
				menuItem.MenuId = menu.Id
				menuItem.InitialCount = count
				menuItem.AvailableCount = count
				menuItem.CreatedAt = timeNow.Time
				menuItem.UpdatedAt = timeNow.Time
				DB.Create(&menuItem)
				DB.Save(&menuItem)
			}
		}
	}

	allMenuItems := []models.MenuItem{}
	DB.Model(menu).Related(&allMenuItems)
	menu.Items = allMenuItems

	return menu, resultError
}

func (c ProviderController) getOrderedDishesByMenuItems(menuItems []models.MenuItem) (orderedDishes map[int64]int64) {

	orderedDishes = make(map[int64]int64)

	for index := range menuItems {

		dishId := menuItems[index].DishId
		initialCount := menuItems[index].InitialCount
		currentCount := menuItems[index].AvailableCount
		orderedCount := initialCount - currentCount
		if orderedCount > 0 {
			orderedDishes[dishId] = orderedCount
		}
	}

	return
}
