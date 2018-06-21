package controllers

import (
	"github.com/revel/revel"
	"lunchapi/app/models"
	"net/http"
	"lunchapi/app/errors"
	"lunchapi/app/responses"
	"lunchapi/app/requests"
	"lunchapi/app/helpers"
)

type NotificationController struct {
	*revel.Controller
}

// @Summary Get Notification
// @Description Get Notifications List
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Notification
// @Success 401 {object} errors.RequestError
// @Router /notifications/index [get]
// @Security Authorization
// @Tags Notifications
func (c NotificationController) Index() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)

	notifications := []models.Notification{}

	query := DB

	if user.Role.Name == "provider" {
		query = query.Where("owner_id = ?", user.Id)
	}
	if user.Role.Name == "master" {
		query = query.
			Where("id IN (?)",
			DB.
				Model(models.UserNotification{}).
				Select("distinct(notification_id)").Where("user_id = ?", user.Id).
				QueryExpr()).
			Preload("UserNotes", "user_id = ?", user.Id)
	}

	query.Find(&notifications)

	return c.RenderJSON(notifications)
}

// @Summary Save Notification
// @Description Add or update your Notification
// @Accept  json
// @Produce  json
// @Param body body requests.NotificationRequest true "Notification Details"
// @Success 200 {object} models.Notification
// @Success 400 {object} errors.RequestError
// @Success 401 {object} errors.RequestError
// @Success 403 {object} errors.RequestError
// @Router /notifications/create [post]
// @Security Authorization
// @Tags Notifications
func (c NotificationController) Create() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)

	if user.Role.Name == "master" {
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(errors.ErrorForbidden("You have no permissions to create notifications"))
	}

	var notificationData requests.NotificationRequest
	c.Params.BindJSON(&notificationData)

	if helpers.IsEmptyString(notificationData.Message)  {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("Notification message cannot be empty", nil))
	}

	notification := models.Notification{}
	notification.OwnerId = user.Id
	notification.Body = notificationData.Message
	notification.IsApproved = false

	DB.Create(&notification)
	DB.Save(&notification)

	return c.RenderJSON(notification)
}

// @Summary Delete Notification
// @Description Delete Notification By Id
// @Accept  json
// @Produce  json
// @Param id path int true "Notification Id"
// @Success 200 {object} responses.GeneralResponse
// @Success 400 {object} errors.RequestError
// @Success 401 {object} errors.RequestError
// @Router /notifications/{id}/delete [delete]
// @Security Authorization
// @Tags Notifications
func (c NotificationController) Delete() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)

	notification := models.Notification{}

	id := c.Params.Route.Get("id")

	query := DB.
		Where("id = ?", id)

	if user.Role.Name != "admin" {
		query = query.
			Where("owner_id = ?", user.Id)
	}

	query = query.First(&notification)

	if notification.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find notification based on your request"))
	}

	if notification.IsApproved {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("Notification was approved already and cannot be removed", nil))
	}

	DB.Where("id = ?", id).Delete(models.Notification{})

	return c.RenderJSON(responses.SuccessfulResponse("Notification has been successfully removed"))
}

// @Summary Approve Notification
// @Description Approve Notification By Id
// @Accept  json
// @Produce  json
// @Param id path int true "Notification Id"
// @Success 200 {object} responses.GeneralResponse
// @Success 401 {object} errors.RequestError
// @Success 403 {object} errors.RequestError
// @Router /notifications/{id}/approve [put]
// @Security Authorization
// @Tags Notifications
func (c NotificationController) Approve() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)

	if user.Role.Name != "admin" {
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(errors.ErrorForbidden("You have no permissions to approve notifications"))
	}

	notification := models.Notification{}

	id := c.Params.Route.Get("id")

	DB.
		Where("id = ?", id).
		First(&notification)

	if notification.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find notification based on your request"))
	}

	if notification.IsApproved {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("This notification is already approved", nil))
	}

	notification.IsApproved = true
	DB.Save(&notification)


	orders := []models.Order{}
	DB.
		Where("item_id IN (?)",
		DB.
			Model(models.MenuItem{}).
			Select("id").Where("menu_id IN (?)",
			DB.
				Model(models.Menu{}).
				Select("id").Where("provider_id = ?", notification.OwnerId).
				QueryExpr()).
			QueryExpr()).
		Find(&orders)

	usersIds := []int64{}

	for _, order := range orders {
		exists, _ := helpers.InArray(order.UserId, usersIds )
		if !exists {
			usersIds = append(usersIds, order.UserId)

			DB.
				Where("user_id = ?", order.UserId).
				Where("notification_id = ?", notification.Id).
				Delete(&models.UserNotification{})

			userNotification := models.UserNotification{
				UserId: order.UserId,
				NotificationId: notification.Id,
				IsRead: false,
			}
			DB.Create(&userNotification)
			DB.Save(&userNotification)
		}
	}

	return c.RenderJSON(responses.SuccessfulResponse("Notification has been successfully approved"))
}

// @Summary Toggle Mark Notification As Read
// @Description Toggle Mark Notification As Read
// @Accept  json
// @Produce  json
// @Param id path int true "Notification Id"
// @Success 200 {object} responses.GeneralResponse
// @Success 401 {object} errors.RequestError
// @Success 403 {object} errors.RequestError
// @Router /notifications/{id}/mark [put]
// @Security Authorization
// @Tags Notifications
func (c NotificationController) Mark() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)

	if user.Role.Name != "master" {
		c.Response.Status = http.StatusForbidden
		return c.RenderJSON(errors.ErrorForbidden("You have no permissions to toggle notification mark"))
	}

	userNotification := models.UserNotification{}

	id := c.Params.Route.Get("id")

	DB.
		Where("user_id = ?", user.Id).
		Where("notification_id = ?", id).
		First(&userNotification)

	if userNotification.Id == 0 {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("This notification is not assigned to you", nil))
	}

	userNotification.IsRead = !userNotification.IsRead
	DB.Save(&userNotification)

	readStatus := "not read"
	if userNotification.IsRead {
		readStatus = "read"
	}

	return c.RenderJSON(responses.SuccessfulResponse("Notification has been successfully marked as " + readStatus))
}
