package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bendigiorgio/go-kv/internal/api"
	"github.com/bendigiorgio/go-kv/internal/engine"
)

// Helper function to create a test router
func setupTestRouter() *api.Router {
	store, _ := engine.NewEngine("test_data.db", "test_flushed.db", 1024)
	return api.NewRouter(store, false)
}

// Helper function to make HTTP requests and assert the response status code
func assertHTTPResponse(t *testing.T, method, url string, body io.Reader, expectedStatusCode int) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatusCode {
		t.Errorf("Expected status code %d, got %d", expectedStatusCode, resp.StatusCode)
	}

	return resp
}

// Helper function to parse JSON response body
func parseJSONResponse(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
}

func TestSetKey(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	data := `{"key":"testKey", "value":"testValue"}`
	resp := assertHTTPResponse(t, http.MethodPost, server.URL+"/set", bytes.NewBuffer([]byte(data)), http.StatusOK)
	defer resp.Body.Close()
}

func TestGetKey(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Set a key first
	data := `{"key":"testKey", "value":"testValue"}`
	assertHTTPResponse(t, http.MethodPost, server.URL+"/set", bytes.NewBuffer([]byte(data)), http.StatusOK)

	// Fetch the key
	resp := assertHTTPResponse(t, http.MethodGet, server.URL+"/get?key=testKey", nil, http.StatusOK)
	defer resp.Body.Close()

	var result map[string]string
	parseJSONResponse(t, resp, &result)

	if result["value"] != "testValue" {
		t.Errorf("Expected 'testValue', got '%s'", result["value"])
	}
}

func TestGetNonExistentKey(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Fetch a non-existent key
	resp := assertHTTPResponse(t, http.MethodGet, server.URL+"/get?key=nonExistentKey", nil, http.StatusNotFound)
	defer resp.Body.Close()
}

func TestDeleteKey(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Set a key first
	data := `{"key":"deleteKey", "value":"deleteValue"}`
	assertHTTPResponse(t, http.MethodPost, server.URL+"/set", bytes.NewBuffer([]byte(data)), http.StatusOK)

	// Delete the key
	resp := assertHTTPResponse(t, http.MethodDelete, server.URL+"/delete?key=deleteKey", nil, http.StatusOK)
	defer resp.Body.Close()

	// Try to get the deleted key
	resp = assertHTTPResponse(t, http.MethodGet, server.URL+"/get?key=deleteKey", nil, http.StatusNotFound)
	defer resp.Body.Close()
}

func TestListKeys(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Set multiple keys
	assertHTTPResponse(t, http.MethodPost, server.URL+"/set", bytes.NewBuffer([]byte(`{"key":"key1", "value":"value1"}`)), http.StatusOK)
	assertHTTPResponse(t, http.MethodPost, server.URL+"/set", bytes.NewBuffer([]byte(`{"key":"key2", "value":"value2"}`)), http.StatusOK)

	// Fetch the key list
	resp := assertHTTPResponse(t, http.MethodGet, server.URL+"/list", nil, http.StatusOK)
	defer resp.Body.Close()

	var result map[string]string
	parseJSONResponse(t, resp, &result)

	if len(result) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(result))
	}
}

func TestFlushDatabase(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Set a key
	assertHTTPResponse(t, http.MethodPost, server.URL+"/set", bytes.NewBuffer([]byte(`{"key":"flushKey", "value":"flushValue"}`)), http.StatusOK)

	// Flush the database
	resp := assertHTTPResponse(t, http.MethodPost, server.URL+"/flush", nil, http.StatusOK)
	defer resp.Body.Close()

	// Check that the database is empty
	resp = assertHTTPResponse(t, http.MethodGet, server.URL+"/list", nil, http.StatusOK)
	defer resp.Body.Close()

	var result map[string]string
	parseJSONResponse(t, resp, &result)

	if len(result) != 0 {
		t.Errorf("Expected an empty database after flush, got %d keys", len(result))
	}
}

func TestCompactDatabase(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Set a key
	assertHTTPResponse(t, http.MethodPost, server.URL+"/set", bytes.NewBuffer([]byte(`{"key":"compactKey", "value":"compactValue"}`)), http.StatusOK)

	// Trigger compaction
	resp := assertHTTPResponse(t, http.MethodPost, server.URL+"/compact", nil, http.StatusOK)
	defer resp.Body.Close()

	// Check if the key is still accessible after compaction
	resp = assertHTTPResponse(t, http.MethodGet, server.URL+"/get?key=compactKey", nil, http.StatusOK)
	defer resp.Body.Close()

	var result map[string]string
	parseJSONResponse(t, resp, &result)

	if result["value"] != "compactValue" {
		t.Errorf("Expected 'compactValue', got '%s'", result["value"])
	}
}
