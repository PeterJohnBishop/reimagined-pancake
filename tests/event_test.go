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
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"reimagined-pancake/server/database"
)

func TestEventLifecycleEndToEnd(t *testing.T) {
	gofakeit.Seed(0)

	fmt.Printf("[ Event lifecycle test | Run ID: %s ]\n", gofakeit.UUID())

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := database.ConnectPGDB()
	require.NoError(t, err, "Failed to connect to test database")

	db.CreateTables()

	router := setupTestRouter(db)

	var jwtToken string
	testEmail := gofakeit.Email()
	testPassword := gofakeit.Password(true, true, true, true, false, 12)

	regPayload := map[string]string{
		"name":     gofakeit.Name(),
		"email":    testEmail,
		"password": testPassword,
	}
	regBody, _ := json.Marshal(regPayload)
	reqReg, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(regBody))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	router.ServeHTTP(wReg, reqReg)

	require.Equal(t, http.StatusCreated, wReg.Code)
	var regResponse map[string]interface{}
	err = json.Unmarshal(wReg.Body.Bytes(), &regResponse)
	require.NoError(t, err)
	jwtToken = regResponse["token"].(string)
	require.NotEmpty(t, jwtToken)

	var eventID string
	eventType := gofakeit.Word()

	t.Run("Store Event", func(t *testing.T) {
		payload := map[string]interface{}{
			"event-type": eventType,
			"data": map[string]string{
				"message": gofakeit.Sentence(5),
				"status":  "test_generated",
			},
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/event", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		eventObj := response["event"].(map[string]interface{})
		eventID = eventObj["id"].(string)

		require.NotEmpty(t, eventID)
	})

	t.Run("Get All Events", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/events", nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var events []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &events)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 1)
	})

	t.Run("Get Event By ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/event/"+eventID, nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var event map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &event)
		assert.NoError(t, err)
		assert.Equal(t, eventID, event["id"])
	})

	t.Run("Get Events By Type", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/events/"+eventType, nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var events []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &events)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 1)

		if len(events) > 0 {
			assert.Equal(t, eventType, events[0]["event_type"])
		}
	})

	t.Run("Delete Event", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/event/"+eventID, nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Verify Deletion", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/event/"+eventID, nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
