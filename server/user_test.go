package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"reimagined-pancake/server/database"
	"reimagined-pancake/server/handlers"
	"reimagined-pancake/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupTestRouter(store *database.DBStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.POST("/register", handlers.RegisterHandler(store))
	r.POST("/login", handlers.LoginHandler(store))

	protected := r.Group("/api")
	protected.Use(utils.RequireAuth())
	{
		protected.PUT("/profile", handlers.UpdateProfileHandler(store))
		protected.PUT("/password", handlers.UpdatePasswordHandler(store))
		protected.DELETE("/profile", handlers.DeleteUserHandler(store))
	}

	return r
}

func performRequest(r http.Handler, method, path string, body []byte, token string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestEndToEndUserFlow(t *testing.T) {

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	store, err := database.ConnectPGDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer store.DB.Close()

	err = store.CreateTables()
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	router := setupTestRouter(store)

	var authToken string

	t.Run("Register User", func(t *testing.T) {
		payload := []byte(`{"name": "Test User", "email": "test@example.com", "password": "supersecret123"}`)
		w := performRequest(router, "POST", "/register", payload, "")

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201 Created, got %v. Body: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		authToken = response["token"].(string)

		if authToken == "" {
			t.Fatal("Expected a token in the registration response, but got none")
		}
	})

	t.Run("Login User", func(t *testing.T) {
		payload := []byte(`{"email": "test@example.com", "password": "supersecret123"}`)
		w := performRequest(router, "POST", "/login", payload, "")

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200 OK, got %v", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		authToken = response["token"].(string)
	})

	t.Run("Update Profile", func(t *testing.T) {
		payload := []byte(`{"name": "Updated User", "email": "updated@example.com"}`)
		w := performRequest(router, "PUT", "/api/profile", payload, authToken)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200 OK, got %v. Body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Update Password", func(t *testing.T) {
		payload := []byte(`{"old_password": "supersecret123", "new_password": "newpassword456"}`)
		w := performRequest(router, "PUT", "/api/password", payload, authToken)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200 OK, got %v. Body: %s", w.Code, w.Body.String())
		}

		loginPayload := []byte(`{"email": "updated@example.com", "password": "supersecret123"}`)
		wLogin := performRequest(router, "POST", "/login", loginPayload, "")
		if wLogin.Code != http.StatusUnauthorized {
			t.Fatalf("Expected login with old password to fail with 401, got %v", wLogin.Code)
		}
	})

	t.Run("Delete User", func(t *testing.T) {
		w := performRequest(router, "DELETE", "/api/profile", nil, authToken)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200 OK, got %v", w.Code)
		}

		loginPayload := []byte(`{"email": "updated@example.com", "password": "newpassword456"}`)
		wLogin := performRequest(router, "POST", "/login", loginPayload, "")

		if wLogin.Code != http.StatusUnauthorized {
			t.Fatalf("Expected login for deleted user to fail with 401, got %v", wLogin.Code)
		}
	})
}
