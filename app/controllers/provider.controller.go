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
// @Tags Providers
func (c ProviderController) Index() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

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

// @Summary Get Provider Profile
// @Description Get Provider Profile
// @Accept  json
// @Produce  json
// @Param provider_id path int true "Provider Id"
// @Success 200 {object} models.User
// @Success 401 {object} errors.RequestError
// @Router /provider/{profile_id} [get]
// @Tags Providers
func (c ProviderController) Profile() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

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
// @Success 200 {object} models.Menu
// @Success 401 {object} errors.RequestError
// @Router /provider/{provider_id}/menus [get]
// @Tags Menus
func (c ProviderController) Menus() revel.Result { //TODO: IMPLEMENT THIS
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

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
		Find(&menus)

	return c.RenderJSON(menus)
}

// @Summary Get Provider's Menu by Date
// @Description Get Provider's Menu by Date
// @Accept  json
// @Produce  json
// @Param provider_id path int true "Provider Id"
// @Param date path string true "Menu Date"
// @Success 200 {array} models.Menu
// @Success 401 {object} errors.RequestError
// @Router /provider/{provider_id}/menus/{date} [get]
// @Tags Menus
func (c ProviderController) Menu() revel.Result { //TODO: IMPLEMENT THIS
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}


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
// @Tags Menus
func (c ProviderController) SaveMenu() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

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
// @Tags Menus
func (c ProviderController) CloneMenu() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	if err := ProviderController.checkProviderPermissions(c); err.Status != 0 {
		return c.RenderJSON(err)
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
// @Tags Menus
func (c ProviderController) DeleteMenu() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

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
// @Tags Dishes
func (c ProviderController) SaveDish() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

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

	}

	DB.Create(&dishData)
	DB.Save(&dishData)

	return c.RenderJSON(dishData.Images)
	return c.RenderJSON(dish.Images)
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
// @Tags Dishes
func (c ProviderController) DeleteDish() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

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

	DB.Where("id = ?", dish.Id).Delete(models.Dish{})

	return c.RenderJSON(responses.SuccessfulResponse("Dish has been successfully removed"))
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

	if user.Id != provider.Id {
		c.Response.Status = http.StatusForbidden
		return errors.ErrorForbidden("You have no permissions to change any data of this user")
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

	differenceInDays := newDateParsed.DiffInDays(dateParsed, false)

	DB.
		Where("provider_id = ?", provider.Id).
		Where("date = ?", dateParsed.DateString()).
		Preload("Items").
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
		return result, errors.ErrorBadRequest("Wrong Deadline date format "+menu.DeadlineAt+errEE.Error(), nil)
	}

	result.DeliveryTime = menu.DeliveryTime
	result.Deadline = deadlineParsed.Copy().AddDays(int(differenceInDays)).DateTimeString()

	counts := []requests.ObjectCounter{}

	for index := range menu.Items {
		dishId := menu.Items[index].DishId
		initialCount := menu.Items[index].InitialCount
		counts = append(counts, requests.ObjectCounter{Id: dishId, Count: initialCount})
	}

	result.Items = counts

	return result, resultError
}

func (c ProviderController) updateMenuData(date string, menuItemsData requests.MenuUpdateRequest) (models.Menu, errors.RequestError) {

	provider := AuthGetCurrentUser(c.Request)

	menu := models.Menu{}
	resultError := errors.RequestError{}

	dateParsed, dateErr := carbon.Parse(carbon.DateFormat, date, provider.Timezone)

	if dateErr != nil {
		c.Response.Status = http.StatusNotFound
		return menu, errors.ErrorNotFound("Unable to find menu based on this date")
	}
	deadlineParsed, deadlineErr := carbon.Parse(carbon.DefaultFormat, menuItemsData.Deadline, provider.Timezone)
	timeParsed, timeErr := carbon.Parse(carbon.TimeFormat, menuItemsData.DeliveryTime, provider.Timezone)
	if (deadlineErr != nil) || (timeErr != nil) {
		c.Response.Status = http.StatusBadRequest
		return menu, errors.ErrorBadRequest("Wrong delivery time or deadline was set", nil)
	}

	DB.
		Where("provider_id = ?", provider.Id).
		Where("date = ?", dateParsed.DateString()).
		First(&menu)

	menuIsNew := false

	if menu.Id == 0 {

		timeNow, _ := carbon.NowInLocation(provider.Timezone)

		menu.ProviderId = provider.Id
		menu.Date = dateParsed.DateString()
		menu.CreatedAt = timeNow.Time
		menu.UpdatedAt = timeNow.Time
		DB.Create(&menu)
		DB.Save(&menu)

		menuIsNew = true
	}

	menu.DeadlineAt = deadlineParsed.DateTimeString()
	menu.DeliveryTime = timeParsed.TimeString()
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
