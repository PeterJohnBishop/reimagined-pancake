package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"reimagined-pancake/server"
	"reimagined-pancake/server/database"
	"reimagined-pancake/server/utils"
)

// a happy path test of User routes, lifecycle, and authentication

func setupTestRouter(db *database.DBStore) *gin.Engine {

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	server.AddOpenRoutes(r, db)

	protected := r.Group("/api")
	protected.Use(utils.RequireAuth())
	server.AddProtectedRoutes(protected, db)

	return r
}

func TestUserLifecycleEndToEnd(t *testing.T) {
	gofakeit.Seed(0)

	fmt.Printf("[ User lifecycle test | Run ID: %s ]\n", gofakeit.UUID())

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := database.ConnectPGDB()
	require.NoError(t, err, "Failed to connect to test database")

	db.CreateTables()

	router := setupTestRouter(db)

	var jwtToken string
	var userID string

	testName := gofakeit.Name()
	testEmail := gofakeit.Email()
	initialPassword := gofakeit.Password(true, true, true, true, false, 12)
	updatedPassword := gofakeit.Password(true, true, true, true, false, 12)

	t.Run("Register User", func(t *testing.T) {
		payload := map[string]string{
			"name":     testName,
			"email":    testEmail,
			"password": initialPassword,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		jwtToken = response["token"].(string)
		userObj := response["user"].(map[string]interface{})
		userID = userObj["id"].(string)

		require.NotEmpty(t, jwtToken)
		require.NotEmpty(t, userID)
	})

	t.Run("Login User", func(t *testing.T) {
		payload := map[string]string{
			"email":    testEmail,
			"password": initialPassword,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get User Profile", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/user/"+userID, nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var user map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &user)
		assert.Equal(t, testEmail, user["email"])
		assert.Equal(t, testName, user["name"])
	})

	t.Run("Update Profile", func(t *testing.T) {
		updatedName := gofakeit.Name()
		updatedEmail := gofakeit.Email()

		payload := map[string]string{
			"name":  updatedName,
			"email": updatedEmail,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("PUT", "/api/user", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		testName = updatedName
		testEmail = updatedEmail
	})

	t.Run("Update Password", func(t *testing.T) {
		payload := map[string]string{
			"old_password": initialPassword,
			"new_password": updatedPassword,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("PUT", "/api/user/password", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("6. Delete User", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/user", nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Verify Deletion via Login Attempt", func(t *testing.T) {
		payload := map[string]string{
			"email":    testEmail,
			"password": updatedPassword,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
