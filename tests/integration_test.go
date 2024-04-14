package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	_ "time"

	_ "banner-manager/db"
	_ "banner-manager/internal/models"
	"banner-manager/internal/handlers"
)

// func TestGetUserBanner(t *testing.T) {
// 	if err := db.ConnectPostgresDB(); err != nil {
// 		t.Fatalf("Failed to initialize and open database: %v", err)
// 	}
// 	defer db.ClosePostgresDB()

// 	mockBanner := &models.Banner{
// 		ID:        1,
// 		FeatureID: 1233,
// 		TagID:     4567,
// 		Title:     "Тестовый баннер",
// 		Text:      "Это тестовый баннер для интеграционного теста",
// 		URL:       "https://example.com/test_banner",
// 		IsActive:  true,
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 	}

// 	_, err := handlers.CreateBanner(*mockBanner)
// 	if err != nil {
// 		t.Fatalf("Failed to create mock banner: %v", err)
// 	}

// 	req, err := http.NewRequest("GET", "/banner/1", nil)
// 	if err != nil {
// 		t.Fatalf("Failed to create request: %v", err)
// 	}

// 	rr := httptest.NewRecorder()

// 	handler := http.HandlerFunc(handlers.GetUserBannerHandler)
// 	handler.ServeHTTP(rr, req)

// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
// 	}

// 	// Выводим тело ответа для отладки
// 	t.Logf("Response body: %s", rr.Body.String())

// 	expectedBannerJSON, err := json.Marshal(mockBanner)
// 	if err != nil {
// 		t.Fatalf("Failed to marshal mock banner to JSON: %v", err)
// 	}
// 	if rr.Body.String() != string(expectedBannerJSON) {
// 		t.Errorf("Handler returned unexpected body: got %v, want %v", rr.Body.String(), string(expectedBannerJSON))
// 	}
// }

func TestGetBannerE2E(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8080/banner/1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(handlers.GetUserBannerHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	expectedTitle := "Ожидаемый заголовок"
	if responseBody["title"] != expectedTitle {
		t.Errorf("Unexpected title: got %v, want %v", responseBody["title"], expectedTitle)
	}
}
