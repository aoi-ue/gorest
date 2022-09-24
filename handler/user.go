package handler

import (
	"net/http"
	"time"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"

	log "github.com/sirupsen/logrus"
)

// GetUsers handles jobs for controller.GetUsers
func GetUsers() (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()
	users := []model.User{}
	posts := []model.Post{}
	hobbies := []model.Hobby{}

	if err := db.Find(&users).Error; err != nil {
		log.WithError(err).Error("error code: 1101")
		httpResponse.Result = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if len(users) == 0 {
		httpResponse.Result = "no user found"
		httpStatusCode = http.StatusNotFound
		return
	}

	for j, user := range users {
		db.Model(&posts).Where("id_user = ?", user.UserID).Find(&posts)
		users[j].Posts = posts

		db.Model(&hobbies).Joins("JOIN user_hobbies ON user_hobbies.hobby_hobby_id=hobbies.hobby_id").
			Joins("JOIN users ON users.user_id=user_hobbies.user_user_id").
			Where("users.user_id = ?", user.UserID).
			Find(&hobbies)
		users[j].Hobbies = hobbies
	}

	httpResponse.Result = users
	httpStatusCode = http.StatusOK
	return
}

// GetUser handles jobs for controller.GetUser
func GetUser(id string) (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()
	user := model.User{}
	posts := []model.Post{}
	hobbies := []model.Hobby{}

	if err := db.Where("user_id = ?", id).First(&user).Error; err != nil {
		httpResponse.Result = "user not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	db.Model(&posts).Where("id_user = ?", user.UserID).Find(&posts)
	user.Posts = posts

	db.Model(&hobbies).Joins("JOIN user_hobbies ON user_hobbies.hobby_hobby_id=hobbies.hobby_id").
		Joins("JOIN users ON users.user_id=user_hobbies.user_user_id").
		Where("users.user_id = ?", user.UserID).
		Find(&hobbies)
	user.Hobbies = hobbies

	httpResponse.Result = user
	httpStatusCode = http.StatusOK
	return
}

// CreateUser handles jobs for controller.CreateUser
func CreateUser(userIDAuth uint64, user model.User) (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()
	userFinal := model.User{}

	// does the user have an existing profile
	if err := db.Where("id_auth = ?", userIDAuth).First(&userFinal).Error; err == nil {
		httpResponse.Result = "user profile found, no need to create a new one"
		httpStatusCode = http.StatusForbidden
		return
	}

	// user must not be able to manipulate all fields
	userFinal.FirstName = user.FirstName
	userFinal.LastName = user.LastName
	userFinal.IDAuth = userIDAuth

	tx := db.Begin()
	if err := tx.Create(&userFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1111")
		httpResponse.Result = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	httpResponse.Result = userFinal
	httpStatusCode = http.StatusCreated
	return
}

// UpdateUser handles jobs for controller.UpdateUser
func UpdateUser(userIDAuth uint64, user model.User) (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()
	userFinal := model.User{}

	// does the user have an existing profile
	if err := db.Where("id_auth = ?", userIDAuth).First(&userFinal).Error; err != nil {
		httpResponse.Result = "no user profile found"
		httpStatusCode = http.StatusNotFound
		return
	}

	// user must not be able to manipulate all fields
	userFinal.UpdatedAt = time.Now()
	userFinal.FirstName = user.FirstName
	userFinal.LastName = user.LastName

	tx := db.Begin()
	if err := tx.Save(&userFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1121")
		httpResponse.Result = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	httpResponse.Result = userFinal
	httpStatusCode = http.StatusOK
	return
}

// AddHobby handles jobs for controller.AddHobby
func AddHobby(userIDAuth uint64, hobby model.Hobby) (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()
	user := model.User{}
	hobbyNew := model.Hobby{}
	hobbyFound := 0 // default (do not create new hobby) = 0, create new hobby = 1

	// does the user have an existing profile
	if err := db.Where("id_auth = ?", userIDAuth).First(&user).Error; err != nil {
		httpResponse.Result = "no user profile found"
		httpStatusCode = http.StatusForbidden
		return
	}

	if err := db.Where("hobby = ?", hobby.Hobby).First(&hobbyNew).Error; err != nil {
		hobbyFound = 1 // create new hobby
	}

	if hobbyFound == 1 {
		hobbyNew.Hobby = hobby.Hobby
		tx := db.Begin()
		if err := tx.Create(&hobbyNew).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1131")
			httpResponse.Result = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		tx.Commit()
		hobbyFound = 0
	}

	if hobbyFound == 0 {
		user.Hobbies = append(user.Hobbies, hobbyNew)
		tx := db.Begin()
		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1132")
			httpResponse.Result = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		tx.Commit()
	}

	httpResponse.Result = user
	httpStatusCode = http.StatusOK
	return
}
