package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupSimpleTestServer() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

func TestHealthEndpoint(t *testing.T) {
	router := setupSimpleTestServer()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestAPIStructure(t *testing.T) {
	router := setupSimpleTestServer()
	
	// Add some mock API endpoints to test structure
	v1 := router.Group("/api/v1")
	
	v1.GET("/workflows", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"workflows": []string{}})
	})
	
	v1.GET("/prompts", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"prompts": []string{}})
	})
	
	v1.GET("/context", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"bundles": []string{}})
	})

	// Test workflows endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/workflows", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Test prompts endpoint
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/prompts", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Test context endpoint
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/context", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestJSONRequestParsing(t *testing.T) {
	router := setupSimpleTestServer()
	
	router.POST("/api/v1/test", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"received": body})
	})

	testData := map[string]interface{}{
		"name": "test",
		"config": map[string]interface{}{
			"enabled": true,
			"count":   42,
		},
	}

	jsonData, _ := json.Marshal(testData)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["received"])
}