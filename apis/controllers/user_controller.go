package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/protengplus/proteng-user-mgmt/models"
	"github.com/protengplus/proteng-user-mgmt/repositories"
	"github.com/protengplus/proteng-user-mgmt/utils/apiutil"
)

type UserController struct {
	userRepository repositories.UserRepository
}

func NewUserController(userRepository repositories.UserRepository) *UserController {
	return &UserController{userRepository: userRepository}
}

// GetAllUsers retrieves all users
func (uc *UserController) GetAllUsers(c *gin.Context) {
	users, err := uc.userRepository.GetAll()
	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return
	}

	// Transform each user using FilteredResponse function
	var filteredUsers []models.UserResponse
	for _, user := range users {
		filteredUsers = append(filteredUsers, models.FilteredResponse(user))
	}

	apiutil.ApiResponseOk(c, filteredUsers)
}

// GetUser retrieves a user by ID
func (uc *UserController) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := uc.userRepository.FindById(id)
	if err != nil {
		apiutil.ApiResponseNotFound(c, err)
		return
	}

	apiutil.ApiResponseOk(c, models.FilteredResponse(user))
}

// UpdateUser updates an existing user
func (uc *UserController) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	user, err := uc.userRepository.FindById(id)
	if err != nil {
		apiutil.ApiResponseNotFound(c, err)
		return
	}
	err = c.BindJSON(&user)
	if err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid request body")
		return
	}

	err = uc.userRepository.Update(id, user)
	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return
	}

	apiutil.ApiResponseOk(c, models.FilteredResponse(user))
}

// DeleteUser deletes a user by ID
func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	err := uc.userRepository.Delete(id)
	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return
	}

	apiutil.ApiResponseOk(c, nil)
}
