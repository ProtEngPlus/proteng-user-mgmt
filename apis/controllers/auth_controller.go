package controllers

import (
	"errors"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"proteng-user-mgmt/models"
	"proteng-user-mgmt/repositories"
	"proteng-user-mgmt/utils"
	"proteng-user-mgmt/utils/apiutil"
)

type AuthController struct {
	userRepository repositories.UserRepository
}

func NewAuthController(userRepository repositories.UserRepository) *AuthController {
	return &AuthController{userRepository: userRepository}
}

func (ac *AuthController) SignInUser(c *gin.Context) {
	var credentials *models.SignInInput

	if err := c.ShouldBindJSON(&credentials); err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid credential")
		return
	}

	user, err := ac.userRepository.FindByEmail(credentials.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid email or password")
			return
		}
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid credential")
		return
	}

	if !slices.Contains(user.Role, credentials.Role) {
		err := errors.New("invalid role")
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid email or password")
		return
	}
	user.Role = []string{credentials.Role}

	if err := utils.VerifyPassword(user.Password, credentials.Password); err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid email or password")
		return
	}

	// Generate Tokens
	duration, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_EXPIRED_IN"))
	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return // Return an error if parsing fails
	}
	accessToken, err := utils.CreateToken(duration, user.Id, credentials.Role, os.Getenv("ACCESS_TOKEN_PRIVATE_KEY"))
	if err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: cannot create token")
		return
	}

	maxAge, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_MAXAGE"))
	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return
	}
	c.SetCookie("access_token", accessToken, maxAge*60, "/", "localhost", false, true)
	c.SetCookie("logged_in", "true", maxAge*60, "/", "localhost", false, false)

	resp := models.FilteredResponse(user)
	resp.AccessToken = accessToken
	resp.CurrentRole = credentials.Role

	apiutil.ApiResponseOk(c, resp)
}
