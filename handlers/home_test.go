package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHome_noToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.GET("/home", Home(nil, "secret"))

	req, _ := http.NewRequest("GET", "/home", nil)
	r.ServeHTTP(w, req)

	// No token = guest response (broken auth - should 401 but doesn't)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "welcome", resp["message"])
}

func TestHome_fakeToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.GET("/home", Home(nil, "secret"))

	req, _ := http.NewRequest("GET", "/home", nil)
	req.Header.Set("Authorization", "Bearer thisisafaketoken")
	r.ServeHTTP(w, req)

	// Fake token still accepted (intentional broken auth)
	assert.Equal(t, http.StatusOK, w.Code)
}
