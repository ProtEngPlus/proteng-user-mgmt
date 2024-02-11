package controllers

import (
	"errors"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/models"
	"github.com/protengplus/proteng-user-mgmt/repositories"
	"github.com/protengplus/proteng-user-mgmt/utils"
	"github.com/protengplus/proteng-user-mgmt/utils/apiutil"
)

type AuthController struct {
	userRepository  repositories.UserRepository
	adminRepository repositories.AdminRepository
}

func NewAuthController(userRepository repositories.UserRepository, adminRepository repositories.AdminRepository) *AuthController {
	return &AuthController{userRepository: userRepository, adminRepository: adminRepository}
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

	if err := utils.VerifyPassword(user.Password, credentials.Password); err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid email or password")
		return
	}

	// Generate Tokens
	duration, err := time.ParseDuration(configs.Config.AccessTokenExpiredIn)
	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return // Return an error if parsing fails
	}
	accessToken, err := utils.CreateToken(duration, user.Id, credentials.Role, configs.Config.AccessTokenPrivateKey)
	if err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: cannot create token")
		return
	}

	// maxAge, err := strconv.Atoi(configs.Config.AccessTokenMaxAge)
	// if err != nil {
	// 	apiutil.ApiResponseInternalServerError(c, err)
	// 	return
	// }
	// c.SetCookie("access_token", accessToken, maxAge*60, "/", "localhost", false, true)
	// c.SetCookie("logged_in", "true", maxAge*60, "/", "localhost", false, false)

	resp := models.FilteredResponse(user)
	resp.AccessToken = accessToken
	resp.CurrentRole = credentials.Role

	apiutil.ApiResponseOk(c, resp)
}

func (ac *AuthController) SignInAdmin(c *gin.Context) {
	var credentials *models.SignInAdminInput

	if err := c.ShouldBindJSON(&credentials); err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid credential")
		return
	}

	admin, err := ac.adminRepository.FindByEmail(credentials.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid email or password")
			return
		}
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid credential")
		return
	}

	if err := utils.VerifyPassword(admin.Password, credentials.Password); err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid email or password")
		return
	}

	// Generate Tokens
	duration, err := time.ParseDuration(configs.Config.AccessTokenExpiredIn)
	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return // Return an error if parsing fails
	}
	accessToken, err := utils.CreateToken(duration, admin.Id, "admin", configs.Config.AccessTokenPrivateKey)
	if err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: cannot create token")
		return
	}

	// maxAge, err := strconv.Atoi(configs.Config.AccessTokenMaxAge)
	// if err != nil {
	// 	apiutil.ApiResponseInternalServerError(c, err)
	// 	return
	// }
	// c.SetCookie("access_token", accessToken, maxAge*60, "/", "localhost", false, true)
	// c.SetCookie("logged_in", "true", maxAge*60, "/", "localhost", false, false)

	resp := models.FilteredAdminResponse(admin)
	resp.AccessToken = accessToken

	apiutil.ApiResponseOk(c, resp)
}
