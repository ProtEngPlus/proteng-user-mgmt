package controllers

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"proteng-user-mgmt/models"
	"proteng-user-mgmt/repositories"
	"proteng-user-mgmt/utils"
)

type AuthController struct {
	userRepository repositories.UserRepository
}

func NewAuthController(userRepository repositories.UserRepository) *AuthController {
	return &AuthController{userRepository: userRepository}
}

func (ac *AuthController) SignInUser(ctx *gin.Context) {
	var credentials *models.SignInInput

	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	user, err := ac.userRepository.FindByEmail(credentials.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid email or password"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := utils.VerifyPassword(user.Password, credentials.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid email or Password"})
		return
	}

	// Generate Tokens
	duration, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_EXPIRED_IN"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Error parsing duration"})
		return // Return an error if parsing fails
	}
	accessToken, err := utils.CreateToken(duration, user.Id, os.Getenv("ACCESS_TOKEN_PRIVATE_KEY"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	maxAge, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_MAXAGE"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Error parsing maxage"})
		return
	}
	ctx.SetCookie("access_token", accessToken, maxAge*60, "/", "localhost", false, true)
	ctx.SetCookie("logged_in", "true", maxAge*60, "/", "localhost", false, false)

	ctx.JSON(http.StatusOK, gin.H{"access_token": accessToken, "user": user})
}
