package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"proteng-user-mgmt/models"
	"proteng-user-mgmt/repositories"
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetUser retrieves a user by ID
func (uc *UserController) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := uc.userRepository.FindById(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser creates a new user
func (uc *UserController) CreateUser(c *gin.Context) {
	var user models.User
	err := c.BindJSON(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = uc.userRepository.Create(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser updates an existing user
func (uc *UserController) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	user, err := uc.userRepository.FindById(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	err = c.BindJSON(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = uc.userRepository.Update(id, user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user by ID
func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	err := uc.userRepository.Delete(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
