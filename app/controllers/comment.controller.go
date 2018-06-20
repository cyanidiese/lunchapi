package controllers

import (
	"github.com/revel/revel"
	"lunchapi/app/models"
	"net/http"
	"lunchapi/app/errors"
	"lunchapi/app/responses"
	"lunchapi/app/helpers"
)

type CommentController struct {
	*revel.Controller
}



// @Summary Get Comments
// @Description Get Comments List
// @Accept  json
// @Produce  json
// @Param dish_id query int false "Dish ID"
// @Param item_id query int false "Menu Item ID"
// @Param owner_id query int false "Owner ID"
// @Success 200 {array} models.Comment
// @Success 401 {object} errors.RequestError
// @Router /comments/index [get]
// @Tags Comments
func (c CommentController) Index() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	comments := []models.Comment{}

	dishId := c.Params.Query.Get("dish_id")
	itemId := c.Params.Query.Get("item_id")
	ownerId := c.Params.Query.Get("owner_id")

	query := DB.
		Preload("Owner")

	if len(dishId) > 0 {
		query = query.Where("dish_id = ?", dishId)
	}
	if len(itemId) > 0 {
		query = query.Where("item_id = ?", itemId)
	}
	if len(ownerId) > 0 {
		query = query.Where("owner_id = ?", ownerId)
	}

	query.Find(&comments)

	return c.RenderJSON(comments)
}

// @Summary Save Comment
// @Description Add or update your Comment
// @Accept  json
// @Produce  json
// @Param body body models.Comment true "Comment Details"
// @Success 200 {object} models.Comment
// @Success 400 {object} errors.RequestError
// @Success 401 {object} errors.RequestError
// @Success 403 {object} errors.RequestError
// @Router /comments/save [post]
// @Tags Comments
func (c CommentController) Save() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	var commentData models.Comment
	c.Params.BindJSON(&commentData)

	if helpers.IsEmptyString(commentData.Body)  {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJSON(errors.ErrorBadRequest("Comment body cannot be empty", nil))
	}

	user := AuthGetCurrentUser(c.Request)

	comment := models.Comment{}

	query := DB.
		Where("id = ?", commentData.Id)

	if user.Role.Name != "admin" {
		query = query.Where("owner_id = ?", user.Id)
	}

	query = query.First(&comment)

	if user.Role.Name == "admin" {
		comment.OwnerId = commentData.OwnerId
	} else {
		comment.OwnerId = user.Id
	}

	comment.ItemId = commentData.ItemId
	comment.DishId = commentData.DishId
	comment.RepliedId = commentData.RepliedId
	comment.Body = commentData.Body

	if comment.Id == 0 {
		DB.Create(&comment)
	}
	DB.Save(&comment)

	return c.RenderJSON(comment)
}

// @Summary Delete Comment
// @Description Delete Comment By Id
// @Accept  json
// @Produce  json
// @Param id path int true "Comment Id"
// @Success 200 {array} responses.GeneralResponse
// @Success 401 {object} errors.RequestError
// @Router /comments/{id}/delete [delete]
// @Tags Comments
func (c CommentController) Delete() revel.Result {
	//Deny Unauthorized users
	if authorized := AuthCheck(c.Request); !authorized {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJSON(errors.ErrorUnauthorized(""))
	}

	user := AuthGetCurrentUser(c.Request)

	comment := models.Comment{}

	id := c.Params.Route.Get("id")

	query := DB.
		Where("id = ?", id)

	if user.Role.Name != "admin" {
		query = query.
			Where("owner_id = ?", user.Id)
	}

	query = query.First(&comment)

	if comment.Id == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJSON(errors.ErrorNotFound("Unable to find comment based on your request"))
	}

	DB.Where("id = ?", id).Delete(models.Comment{})

	return c.RenderJSON(responses.SuccessfulResponse("Comment has been successfully removed"))
}
