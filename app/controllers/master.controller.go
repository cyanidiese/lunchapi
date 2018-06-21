package controllers

import (
	"github.com/revel/revel"
	"net/http"
	"lunchapi/app/errors"
	"lunchapi/app/structs"
	"lunchapi/app/models"
	"lunchapi/app/requests"
	"github.com/uniplaces/carbon"
	"strings"
	"lunchapi/app/responses"
	"github.com/revel/revel/cache"
	"time"
)

type MasterController struct {
	*revel.Controller
}

// @Summary Master Info
// @Description Get Login Info
// @Accept  json
// @Produce  json
// @Success 200 {object} models.User
// @Success 401 {object} errors.RequestError
// @Router /master/index [get]
// @Security Authorization
// @Tags Master
func (c MasterController) Index() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)

	return c.RenderJSON(user)
}

// @Summary Master Update
// @Description Update Master
// @Accept  json
// @Produce  json
// @Param body  body requests.UpdateProfileRequest true "Profile Details"
// @Success 200 {object} models.User
// @Success 401 {object} errors.RequestError
// @Router /master/update [post]
// @Security Authorization
// @Tags Master
func (c MasterController) Update() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)

	var requestBody requests.UpdateProfileRequest
	c.Params.BindJSON(&requestBody)

	appliedChanges, resultError, respStatus := UpdateUserData(requestBody, &user, false)
	if respStatus != 0 {
		c.Response.Status = respStatus
		return c.RenderJSON(resultError)
	}

	if appliedChanges {

		DB.Save(&user)

		cacheKey := "user_" + user.Token
		DB.
			Where("`is_disabled` != 1").
			Where("`token` = ?", user.Token).
			Preload("Image").
			Preload("Role").
			First(&user)

		go cache.Set(cacheKey, user, 30*time.Minute)

	}

	return c.RenderJSON(user)
}

// @Summary Master Stats
// @Description Get Master Stats
// @Accept  json
// @Produce  json
// @Param from_date path string true "Start Date"
// @Param to_date path string true "End Date"
// @Success 200 {object} responses.StatsResponse
// @Success 400 {object} errors.RequestError
// @Success 401 {object} errors.RequestError
// @Router /master/stats/{from_date}/to/{to_date} [get]
// @Security Authorization
// @Tags Master
func (c MasterController) Stats() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	master := AuthGetCurrentUser(c.Request)

	fromDate := c.Params.Route.Get("from_date")
	toDate := c.Params.Route.Get("to_date")

	fromDateParsed, fromDateErr := carbon.Parse(carbon.DateFormat, fromDate, master.Timezone)
	toDateParsed, toDateErr := carbon.Parse(carbon.DateFormat, toDate, master.Timezone)

	if (fromDateErr != nil) || (toDateErr != nil) {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("Wrong date range", nil))
	}

	orders := []models.Order{}
	DB.
		Where("user_id = ? ", master.Id).
		Where("created_at > ? ", fromDateParsed.StartOfDay().DateTimeString()).
		Where("updated_at < ? ", toDateParsed.EndOfDay().DateTimeString()).
		Preload("MenuItem").
		Preload("MenuItem.Dish").
		Find(&orders)

	var weight, calories, price float64 = 0, 0, 0
	for _, order := range orders {
		weight += order.MenuItem.Dish.Weight
		calories += order.MenuItem.Dish.Calories
		price += order.MenuItem.Price
	}

	result := responses.StatsResponse{
		Weight: weight,
		Calories: calories,
		Price: price,
	}

	return c.RenderJSON(result)
}

// @Summary Master Orders History
// @Description Get Master Orders History
// @Accept  json
// @Produce  json
// @Param from_date path string true "Start Date"
// @Param to_date path string true "End Date"
// @Success 200 {array} models.Order
// @Success 400 {object} errors.RequestError
// @Success 401 {object} errors.RequestError
// @Router /master/history/{from_date}/to/{to_date} [get]
// @Security Authorization
// @Tags Master
func (c MasterController) History() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	master := AuthGetCurrentUser(c.Request)

	fromDate := c.Params.Route.Get("from_date")
	toDate := c.Params.Route.Get("to_date")

	fromDateParsed, fromDateErr := carbon.Parse(carbon.DateFormat, fromDate, master.Timezone)
	toDateParsed, toDateErr := carbon.Parse(carbon.DateFormat, toDate, master.Timezone)

	if (fromDateErr != nil) || (toDateErr != nil) {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("Wrong date range", nil))
	}

	orders := []models.Order{}
	DB.
		Where("user_id = ? ", master.Id).
		Where("created_at > ? ", fromDateParsed.StartOfDay().DateTimeString()).
		Where("updated_at < ? ", toDateParsed.EndOfDay().DateTimeString()).
		Preload("MenuItem").
		Preload("MenuItem.Dish").
		Preload("MenuItem.Dish.Name").
		Preload("MenuItem.Dish.Description").
		Preload("MenuItem.Dish.Images").
		Find(&orders)

	return c.RenderJSON(orders)
}

// @Summary Orders List By Date
// @Description Orders List By Date
// @Accept  json
// @Produce  json
// @Param date path string true "Date"
// @Param provider_id path int true "Provider Id"
// @Success 200 {array} models.Order
// @Success 401 {object} errors.RequestError
// @Router /master/orders/{provider_id}/{date} [get]
// @Security Authorization
// @Tags Orders
func (c MasterController) Orders() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	providerId := c.Params.Route.Get("provider_id")
	menuDate := c.Params.Route.Get("date")

	provider, err := MasterController.checkProviderIsAvailable(c, providerId)
	if err.Status != 0 {
		return c.RenderJSON(err)
	}

	menu, err := MasterController.checkMenuIsAvailable(c, menuDate, provider)
	if err.Status != 0 {
		return c.RenderJSON(err)
	}

	master := AuthGetCurrentUser(c.Request)
	items := []models.MenuItem{}
	orders := []models.Order{}
	//order := models.Order{}
	DB.Where("menu_id = ? ", menu.Id).Find(&items)
	itemsIds := []int64{}
	for _, val := range items {
		itemsIds = append(itemsIds, val.Id)
	}
	DB.Where("user_id = ? ", master.Id).Where("item_id IN (?) ", itemsIds).Find(&orders)

	return c.RenderJSON(orders)
}

// @Summary Make Order
// @Description Make Order for some Menu Items
// @Accept  json
// @Produce  json
// @Param date        path string true "Date"
// @Param provider_id path int    true "Provider Id"
// @Param body        body requests.OrderingRequest true "Menu Items Ids with count ordering"
// @Success 200 {array} models.Order
// @Success 401 {object} errors.RequestError
// @Router /master/orders/{provider_id}/{date}/make [post]
// @Security Authorization
// @Tags Orders
func (c MasterController) MakeOrder() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	providerId := c.Params.Route.Get("provider_id")
	menuDate := c.Params.Route.Get("date")

	provider, err := MasterController.checkProviderIsAvailable(c, providerId)
	if err.Status != 0 {
		return c.RenderJSON(err)
	}

	menu, err := MasterController.checkMenuIsBeforeDeadline(c, menuDate, provider)
	if err.Status != 0 {
		return c.RenderJSON(err)
	}

	master := AuthGetCurrentUser(c.Request)

	var requestBody requests.OrderingRequest
	c.Params.BindJSON(&requestBody)

	items := []models.MenuItem{}

	itemsIds := []int64{}

	for _, val := range requestBody.Items {
		itemsIds = append(itemsIds, val.Id)
	}

	DB.Where("menu_id = ?", menu.Id).Where("id IN (?)", itemsIds).Find(&items)

	if len(items) < len(requestBody.Items) {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("There is a difference between items in request and real items", nil))
	}

	availableCounts := MasterController.getAvailableCountsOfMenuItems(c, items)

	for _, item := range requestBody.Items {

		itemsIds := make(map[int64]int64)

		if item.Count > availableCounts[item.Id] {
			itemsIds[item.Id] = availableCounts[item.Id]
		}

		if len(itemsIds) != 0 {
			c.Response.Status = http.StatusBadRequest
			return c.RenderJSON(errors.ErrorBadRequest("There are lower count of some dishes available", itemsIds))
		}
	}

	for _, val := range requestBody.Items {

		item := models.MenuItem{}
		order := models.Order{}
		DB.Where("id = ? ", val.Id).Find(&item)
		DB.Where("user_id = ? ", master.Id).Where("item_id = ? ", val.Id).Find(&order)

		var orderedCount int64 = 0

		// if order is new and user has no orders of this menu item
		if order.Id == 0 {
			if val.Count > 0 {
				orderedCount = val.Count

				order.UserId = master.Id
				order.ItemId = val.Id
				order.OrderedCount = val.Count

				DB.Create(&order)
				DB.Save(&order)

			}
		} else {
			//if users are removing their order
			if val.Count == 0 {
				orderedCount = -1 * order.OrderedCount
				DB.Delete(&order)
			} else {

				//if user tries to make order count negative we just set it to 0
				if order.OrderedCount+val.Count < 0 {
					val.Count = -1 * order.OrderedCount
				}

				//if new count is not 0 - it is difference
				orderedCount = val.Count
				order.OrderedCount = order.OrderedCount + orderedCount
				DB.Save(&order)

				if order.OrderedCount == 0 {
					DB.Delete(&order)
				}
			}
		}

		// decrease available count in this menu item (or increase if difference was negative)
		item.AvailableCount = item.AvailableCount - orderedCount
		DB.Save(&item)
	}

	return MasterController.Orders(c)
}

// @Summary Get Favorites
// @Description Make Order for some Dishes
// @Accept  json
// @Produce  json
// @Success 200 {object} structs.IdsArray
// @Success 401 {object} errors.RequestError
// @Router /master/favorites [get]
// @Security Authorization
// @Tags Favorites
func (c MasterController) Favorites() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	master := AuthGetCurrentUser(c.Request)

	favorites := []models.Favorite{}

	dishesIds := []int64{}

	DB.Where("user_id = ?", master.Id).Find(&favorites)

	for _, val := range favorites {
		dishesIds = append(dishesIds, val.DishId)
	}

	return c.RenderJSON(structs.IdsArray{Ids: dishesIds})
}

// @Summary Add Dish To Favorites
// @Description Add Dish To Favorites
// @Accept  json
// @Produce  json
// @Param body body structs.SimpleId true "Dish Id"
// @Success 200 {object} structs.IdsArray
// @Success 401 {object} errors.RequestError
// @Success 404 {object} errors.RequestError
// @Router /master/favorites/add [post]
// @Security Authorization
// @Tags Favorites
func (c MasterController) AddFavorite() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	var requestBody structs.SimpleId
	c.Params.BindJSON(&requestBody)

	master := AuthGetCurrentUser(c.Request)

	favorite := models.Favorite{}
	dish := models.Dish{}

	DB.Where("dish_id = ?", requestBody.Id).First(&dish)

	if dish.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find dish by id"))
	}

	DB.Where("dish_id = ?", requestBody.Id).Where("user_id = ?", master.Id).First(&favorite)

	if favorite.Id == 0 {
		favorite.UserId = master.Id
		favorite.DishId = requestBody.Id
		DB.Create(&favorite)
		DB.Save(&favorite)
	}

	return MasterController.Favorites(c)
}

// @Summary Remove Dish From Favorites
// @Description Remove Dish From Favorites
// @Accept  json
// @Produce  json
// @Param body body structs.SimpleId true "Dish Id"
// @Success 200 {object} structs.IdsArray
// @Success 401 {object} errors.RequestError
// @Router /master/favorites/remove [delete]
// @Security Authorization
// @Tags Favorites
func (c MasterController) RemoveFavorite() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	var requestBody structs.SimpleId
	c.Params.BindJSON(&requestBody)

	master := AuthGetCurrentUser(c.Request)

	DB.Where("dish_id = ?", requestBody.Id).Where("user_id = ?", master.Id).Delete(models.Favorite{})

	return MasterController.Favorites(c)
}

func (c MasterController) getAvailableCountsOfMenuItems(menuItems []models.MenuItem) (availableCounts map[int64]int64) {

	availableCounts = make(map[int64]int64)

	for index := range menuItems {
		itemId := menuItems[index].Id
		availableCount := menuItems[index].AvailableCount
		availableCounts[itemId] = availableCount
	}

	return
}

func (c MasterController) checkProviderIsAvailable(providerId string) (provider models.User, resultError errors.RequestError) {

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

func (c MasterController) checkMenuIsAvailable(menuDate string, provider models.User) (menu models.Menu, resultError errors.RequestError) {

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

func (c MasterController) checkMenuIsBeforeDeadline(menuDate string, provider models.User) (menu models.Menu, resultError errors.RequestError) {

	menu, err := MasterController.checkMenuIsAvailable(c, menuDate, provider)
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
