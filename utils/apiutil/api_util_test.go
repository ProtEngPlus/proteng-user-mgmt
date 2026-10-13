package apiutil

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/protengplus/proteng-user-mgmt/models"
)

func TestApiResponseConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name            string
		messages        []string
		expectedMessage string
	}{
		{
			name:            "with message",
			messages:        []string{"error: email already registered"},
			expectedMessage: "error: email already registered",
		},
		{
			name:            "without message",
			messages:        nil,
			expectedMessage: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)

			ApiResponseConflict(c, errors.New("duplicate email"), tc.messages...)

			if recorder.Code != http.StatusConflict {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusConflict)
			}

			var body models.HttpResponseError
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("could not unmarshal response body: %v", err)
			}

			if body.Code != http.StatusConflict {
				t.Errorf("body code = %d, want %d", body.Code, http.StatusConflict)
			}
			if body.Error != "duplicate email" {
				t.Errorf("body error = %q, want %q", body.Error, "duplicate email")
			}
			if body.Message != tc.expectedMessage {
				t.Errorf("body message = %q, want %q", body.Message, tc.expectedMessage)
			}
		})
	}
}
