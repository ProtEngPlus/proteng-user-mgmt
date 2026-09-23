package controllers

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/database"
	"github.com/protengplus/proteng-user-mgmt/models"
	"github.com/protengplus/proteng-user-mgmt/repositories"
	"github.com/protengplus/proteng-user-mgmt/utils"
	"github.com/protengplus/proteng-user-mgmt/utils/apiutil"
)

type AuthController struct {
	userRepository  repositories.UserRepository
	adminRepository repositories.AdminRepository
	collection      *mongo.Collection
	temp            *template.Template
}

func NewAuthController(userRepository repositories.UserRepository, adminRepository repositories.AdminRepository, temp *template.Template) *AuthController {
	return &AuthController{
		userRepository:  userRepository,
		adminRepository: adminRepository,
		collection:      database.GetCollection("users"),
		temp:            temp,
	}
}

const verificationTokenDaysTTL = 7

func (ac *AuthController) RegisterUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid request body")
		return
	}

	verificationToken := randstr.String(20)
	user.IsVerified = false
	user.EmailVerificationToken = utils.Encode(verificationToken)
	user.EmailVerificationTokenExpire = time.Now().Add(verificationTokenDaysTTL * 24 * time.Hour)

	if err := ac.userRepository.Create(&user); err != nil {
		if errors.Is(err, repositories.ErrDuplicateEmail) {
			apiutil.ApiResponseConflict(c, err, "error: email already registered")
			return
		}
		apiutil.ApiResponseInternalServerError(c, err)
		return
	}

	firstName := user.Name
	if strings.Contains(firstName, " ") {
		firstName = strings.Split(firstName, " ")[1]
	}

	emailData := utils.EmailData{
		URL:        configs.Config.Origin + "/success-verified?token=" + verificationToken,
		FirstName:  firstName,
		Subject:    fmt.Sprintf("Your email verification token (valid for %d days)", verificationTokenDaysTTL),
		ExpiryDays: verificationTokenDaysTTL,
	}

	if err := utils.SendEmail(&user, &emailData, ac.temp, "verificationEmail"); err != nil {
		apiutil.ApiResponseBadGateway(c, err, "error: account created but verification email could not be sent, please resend from the verification page")
		return
	}

	apiutil.ApiResponseOk(c, models.FilteredResponse(&user), "Verification email sent")
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

	if !user.IsVerified {
		err := errors.New("email not verified")
		apiutil.ApiResponseErrorBadRequest(c, err, "error: email not verified")
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

func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var userCredential *models.ForgotPasswordInput

	if err := c.ShouldBindJSON(&userCredential); err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid credential")
		return
	}

	message := "You will receive a reset email if user with that email exist"

	user, err := ac.userRepository.FindByEmail(userCredential.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			apiutil.ApiResponseOk(c, userCredential, message)
			return
		}
		apiutil.ApiResponseBadGateway(c, err)
		return
	}

	if !user.IsVerified {
		err := errors.New("email not verified")
		apiutil.ApiResponseErrorBadRequest(c, err, "error: email not verified")
		return
	}

	// Generate Verification Code
	resetToken := randstr.String(20)

	passwordResetToken := utils.Encode(resetToken)

	// Update User in Database
	query := bson.D{{Key: "email", Value: utils.NormalizeEmail(userCredential.Email)}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "passwordResetToken", Value: passwordResetToken}, {Key: "passwordResetTokenExpire", Value: time.Now().Add(time.Minute * 15)}}}}
	result, err := ac.collection.UpdateOne(context.Background(), query, update)

	if result.MatchedCount == 0 {
		apiutil.ApiResponseBadGateway(c, err, "There was an error sending email")
		return
	}

	if err != nil {
		apiutil.ApiResponseForbidden(c, err)
		return
	}
	var firstName = user.Name

	if strings.Contains(firstName, " ") {
		firstName = strings.Split(firstName, " ")[1]
	}

	// Send Email
	emailData := utils.EmailData{
		URL:       configs.Config.Origin + "/reset-password?token=" + resetToken,
		FirstName: firstName,
		Subject:   "Your password reset token (valid for 10 minutes)",
	}

	err = utils.SendEmail(user, &emailData, ac.temp, "resetPassword.html")
	if err != nil {
		apiutil.ApiResponseBadGateway(c, err, "There was an error sending email")
		return
	}
	apiutil.ApiResponseOk(c, userCredential, message)
}

func (ac *AuthController) ResetPassword(c *gin.Context) {
	resetToken := c.Params.ByName("resetToken")
	var userCredential *models.ResetPasswordInput

	if err := c.ShouldBindJSON(&userCredential); err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid credential")
		return
	}

	hashedPassword, _ := utils.HashPassword(userCredential.Password)

	passwordResetToken := utils.Encode(resetToken)

	// Update User in Database
	query := bson.D{{Key: "passwordResetToken", Value: passwordResetToken}, {Key: "passwordResetTokenExpire", Value: bson.D{{Key: "$gt", Value: time.Now()}}}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "password", Value: hashedPassword}}}, {Key: "$unset", Value: bson.D{{Key: "passwordResetToken", Value: ""}, {Key: "passwordResetTokenExpire", Value: ""}}}}
	result, err := ac.collection.UpdateOne(context.Background(), query, update)

	if result.MatchedCount == 0 {
		apiutil.ApiResponseErrorBadRequest(c, fmt.Errorf("invalid token"), "Token is invalid or has expired")
		return
	}

	if err != nil {
		apiutil.ApiResponseForbidden(c, err)
		return
	}

	// c.SetCookie("access_token", "", -1, "/", "localhost", false, true)
	// c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	// c.SetCookie("logged_in", "", -1, "/", "localhost", false, true)

	apiutil.ApiResponseOk(c, nil, "Password data updated successfully")
}

func (ac *AuthController) ChangePassword(c *gin.Context) {
	id := c.Param("id")
	user, err := ac.userRepository.FindById(id)
	if err != nil {
		apiutil.ApiResponseNotFound(c, err)
		return
	}
	var userCredential *models.ChangePasswordInput
	if err := c.ShouldBindJSON(&userCredential); err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid credential")
		return
	}

	// Validate Current Password
	if err := utils.VerifyPassword(user.Password, userCredential.CurrentPassword); err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: incorrect current password")
		return
	}

	// Update User in Database
	hashedPassword, _ := utils.HashPassword(userCredential.NewPassword)

	query := bson.D{{Key: "email", Value: utils.NormalizeEmail(user.Email)}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "password", Value: hashedPassword}}}}
	result, err := ac.collection.UpdateOne(context.Background(), query, update)

	if result.MatchedCount == 0 {
		apiutil.ApiResponseErrorBadRequest(c, err, "Cannot update password")
		return
	}

	if err != nil {
		apiutil.ApiResponseForbidden(c, err)
		return
	}

	apiutil.ApiResponseOk(c, nil, "Password data updated successfully")
}

func (ac *AuthController) SendVerification(c *gin.Context) {
	var userCredential *models.SendVerificationInput

	if err := c.ShouldBindJSON(&userCredential); err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid credential")
		return
	}

	message := "You will receive a verification email if user with that email exists."

	user, err := ac.userRepository.FindByEmail(userCredential.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			apiutil.ApiResponseOk(c, userCredential, message)
			return
		}
		apiutil.ApiResponseBadGateway(c, err)
		return
	}

	// Can resent email after passing 60s
	lastSent := user.EmailVerificationTokenExpire.Add(-verificationTokenDaysTTL * 24 * time.Hour)
	if time.Since(lastSent) < 60*time.Second {
		apiutil.ApiResponseOk(c, userCredential, message)
		return
	}

	// Generate Verification Code
	verificationToken := randstr.String(20)

	emailVerificationToken := utils.Encode(verificationToken)

	// Update User in Database
	query := bson.D{{
		Key:   "email",
		Value: utils.NormalizeEmail(userCredential.Email),
	}}
	update := bson.D{{
		Key: "$set",
		Value: bson.D{
			{
				Key:   "emailVerificationToken",
				Value: emailVerificationToken,
			},
			{
				Key:   "emailVerificationTokenExpire",
				Value: time.Now().Add(verificationTokenDaysTTL * 24 * time.Hour),
			},
		}}}
	result, err := ac.collection.UpdateOne(context.Background(), query, update)

	if result.MatchedCount == 0 {
		apiutil.ApiResponseBadGateway(c, err, "There was an error sending email")
		return
	}

	if err != nil {
		apiutil.ApiResponseForbidden(c, err)
		return
	}
	var firstName = user.Name

	if strings.Contains(firstName, " ") {
		firstName = strings.Split(firstName, " ")[1]
	}

	// Send Email
	emailData := utils.EmailData{
		URL:        configs.Config.Origin + "/success-verified?token=" + verificationToken,
		FirstName:  firstName,
		Subject:    fmt.Sprintf("Your email verification token (valid for %d days)", verificationTokenDaysTTL),
		ExpiryDays: verificationTokenDaysTTL,
	}

	err = utils.SendEmail(user, &emailData, ac.temp, "verificationEmail")
	if err != nil {
		apiutil.ApiResponseBadGateway(c, err, "There was an error sending email")
		return
	}
	apiutil.ApiResponseOk(c, userCredential, message)
}

func (ac *AuthController) VerifyEmail(c *gin.Context) {
	verificationToken := c.Params.ByName("verificationToken")
	emailVerificationToken := utils.Encode(verificationToken)

	// Verify token in one atomic operation
	query := bson.D{
		{Key: "emailVerificationToken", Value: emailVerificationToken},
		{Key: "emailVerificationTokenExpire", Value: bson.D{{Key: "$gt", Value: time.Now()}}},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "is_verified", Value: true}}},
		{Key: "$unset", Value: bson.D{
			{Key: "emailVerificationToken", Value: ""},
			{Key: "emailVerificationTokenExpire", Value: ""},
		}},
	}

	var user models.User
	err := ac.collection.FindOneAndUpdate(
		context.Background(), query, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			apiutil.ApiResponseErrorBadRequest(c, fmt.Errorf("invalid token"), "Token is invalid or has expired")
			return
		}
		apiutil.ApiResponseForbidden(c, err)
		return
	}

	// Generate Tokens
	var duration time.Duration
	duration, err = time.ParseDuration(configs.Config.AccessTokenExpiredIn)

	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return // Return an error if parsing fails
	}
	accessToken, err := utils.CreateToken(duration, user.Id, "user", configs.Config.AccessTokenPrivateKey)
	if err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: cannot create token")
		return
	}

	resp := models.FilteredResponse(&user)
	resp.AccessToken = accessToken
	resp.CurrentRole = "user"

	apiutil.ApiResponseOk(c, resp, "Email verified successfully")
}
