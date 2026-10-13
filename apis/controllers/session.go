package controllers

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/models"
	"github.com/protengplus/proteng-user-mgmt/repositories"
	"github.com/protengplus/proteng-user-mgmt/utils"
	"github.com/protengplus/proteng-user-mgmt/utils/apiutil"
)

type sessionTokens struct {
	AccessToken  string
	RefreshToken string
}

func (ac *AuthController) issueSession(subjectId primitive.ObjectID, role string, familyId primitive.ObjectID) (*sessionTokens, error) {
	accessTTL, err := time.ParseDuration(configs.Config.AccessTokenExpiredIn)
	if err != nil {
		return nil, fmt.Errorf("parse access token ttl: %w", err)
	}
	refreshTTL, err := time.ParseDuration(configs.Config.RefreshTokenExpiredIn)
	if err != nil {
		return nil, fmt.Errorf("parse refresh token ttl: %w", err)
	}

	accessToken, err := utils.CreateToken(accessTTL, subjectId, role, configs.Config.AccessTokenPrivateKey)
	if err != nil {
		return nil, err
	}

	raw, hash, err := utils.NewRefreshToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	err = ac.refreshTokenRepository.Create(&models.RefreshToken{
		TokenHash: hash,
		FamilyId:  familyId,
		SubjectId: subjectId,
		Role:      role,
		ExpiresAt: now.Add(refreshTTL),
		CreatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &sessionTokens{AccessToken: accessToken, RefreshToken: raw}, nil
}

func (ac *AuthController) Logout(c *gin.Context) {
	var input models.RefreshTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid request body")
		return
	}

	token, err := ac.refreshTokenRepository.FindByHash(utils.HashRefreshToken(input.RefreshToken))
	if errors.Is(err, repositories.ErrRefreshTokenInvalid) {
		apiutil.ApiResponseOk(c, nil, "Logged out")
		return
	}
	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return
	}

	if err := ac.refreshTokenRepository.RevokeFamily(token.FamilyId, time.Now()); err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return
	}

	apiutil.ApiResponseOk(c, nil, "Logged out")
}
