package apiutil

import (
	"net/http"

	"github.com/protengplus/proteng-user-mgmt/models"

	"github.com/gin-gonic/gin"
)

func ApiResponseOk(c *gin.Context, data interface{}, messages ...string) {
	if len(messages) == 0 {
		messages = append(messages, "")
	}
	c.JSON(http.StatusOK, models.HttpResponseOK{
		Code:    http.StatusOK,
		Data:    data,
		Message: messages[0],
	})
}

func ApiResponseErrorBadRequest(c *gin.Context, err error, messages ...string) {
	if len(messages) == 0 {
		messages = append(messages, "")
	}
	c.JSON(http.StatusBadRequest, models.HttpResponseError{
		Code:    http.StatusBadRequest,
		Error:   err.Error(),
		Message: messages[0],
	})
}

func ApiResponseInternalServerError(c *gin.Context, err error, messages ...string) {
	if len(messages) == 0 {
		messages = append(messages, "")
	}
	c.JSON(http.StatusInternalServerError, models.HttpResponseError{
		Code:    http.StatusInternalServerError,
		Error:   err.Error(),
		Message: messages[0],
	})
}

func ApiResponseNotFound(c *gin.Context, err error, messages ...string) {
	if len(messages) == 0 {
		messages = append(messages, "")
	}
	c.JSON(http.StatusNotFound, models.HttpResponseError{
		Code:    http.StatusNotFound,
		Error:   err.Error(),
		Message: messages[0],
	})
}

func ApiResponseForbidden(c *gin.Context, err error, messages ...string) {
	if len(messages) == 0 {
		messages = append(messages, "")
	}
	c.AbortWithStatusJSON(http.StatusForbidden, models.HttpResponseError{
		Code:    http.StatusForbidden,
		Error:   err.Error(),
		Message: messages[0],
	})
}

func ApiResponseBadGateway(c *gin.Context, err error, messages ...string) {
	if len(messages) == 0 {
		messages = append(messages, "")
	}
	c.AbortWithStatusJSON(http.StatusBadGateway, models.HttpResponseError{
		Code:    http.StatusBadGateway,
		Error:   err.Error(),
		Message: messages[0],
	})
}

func ApiResponseConflict(c *gin.Context, err error, messages ...string) {
	if len(messages) == 0 {
		messages = append(messages, "")
	}
	c.JSON(http.StatusConflict, models.HttpResponseError{
		Code:    http.StatusConflict,
		Error:   err.Error(),
		Message: messages[0],
	})
}
