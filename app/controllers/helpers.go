package controllers

import (
	"lunchapi/app/requests"
	"lunchapi/app/helpers"
	"net/http"
	"lunchapi/app/errors"
	"lunchapi/app/models"
)

func UpdateUserData(userData requests.UpdateProfileRequest, user *models.User, creatingNewUser bool) (appliedChanges bool, resultError errors.RequestError, respStatus int) {

	appliedChanges = false

	if !helpers.IsEmptyString(userData.Password){
		pass, err := AuthHashPassword(userData.Password)
		if err != nil {
			respStatus = http.StatusBadRequest
			resultError = errors.ErrorBadRequest("You cannot use this password. Please choose another one", nil)
			return
		}
		user.Password = pass
		appliedChanges = true
	}

	if !helpers.IsEmptyString(userData.FirstName){
		user.FirstName = userData.FirstName
		appliedChanges = true
	}

	if !helpers.IsEmptyString(userData.LastName){
		user.LastName = userData.LastName
		appliedChanges = true
	}

	if !helpers.IsEmptyString(userData.Alias){
		user.Alias = userData.Alias
		appliedChanges = true
	}

	if !helpers.IsEmptyNumber(userData.ProviderId){
		user.ProviderId = userData.ProviderId
		appliedChanges = true
	}

	if !helpers.IsEmptyNumber(userData.OfficeId){
		user.OfficeId = userData.OfficeId
		appliedChanges = true
	}

	if !helpers.IsEmptyString(userData.Timezone){
		user.Timezone = userData.Timezone
		appliedChanges = true
	}

	if !helpers.IsEmptyString(userData.Language){
		user.Language = userData.Language
		appliedChanges = true
	}

	if creatingNewUser {
		DB.Create(&user)
		DB.Save(&user)
	}

	if !helpers.IsEmptyString(userData.ImageGuid){
		if creatingNewUser || (user.Image.Guid != userData.ImageGuid) {

			if !creatingNewUser {
				DB.Where("id = ?", user.Image.Id).Delete(models.Image{})
			}

			image := models.Image{
				UserId: user.Id,
				Guid: userData.ImageGuid,
			}
			DB.Create(&image)
			DB.Save(&image)
			appliedChanges = true
		}
	}

	return
}
