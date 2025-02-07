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
	store := engine.NewEngine("test_data.txt", "test_flushed.txt", 1024)
	return api.NewRouter(store, false)
}

func TestSetKey(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	data := `{"key":"testKey", "value":"testValue"}`
	resp, err := http.Post(server.URL+"/set", "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestGetKey(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Set a key first
	data := `{"key":"testKey", "value":"testValue"}`
	_, _ = http.Post(server.URL+"/set", "application/json", bytes.NewBuffer([]byte(data)))

	// Fetch the key
	resp, err := http.Get(server.URL + "/get?key=testKey")
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var result map[string]string
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	if result["value"] != "testValue" {
		t.Errorf("Expected 'testValue', got '%s'", result["value"])
	}
}

func TestDeleteKey(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Set a key first
	data := `{"key":"deleteKey", "value":"deleteValue"}`
	_, _ = http.Post(server.URL+"/set", "application/json", bytes.NewBuffer([]byte(data)))

	// Delete the key
	req, _ := http.NewRequest("DELETE", server.URL+"/delete?key=deleteKey", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Try to get the deleted key
	resp, _ = http.Get(server.URL + "/get?key=deleteKey")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found after deletion, got %d", resp.StatusCode)
	}
}

func TestListKeys(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Set multiple keys
	http.Post(server.URL+"/set", "application/json", bytes.NewBuffer([]byte(`{"key":"key1", "value":"value1"}`)))
	http.Post(server.URL+"/set", "application/json", bytes.NewBuffer([]byte(`{"key":"key2", "value":"value2"}`)))

	// Fetch the key list
	resp, err := http.Get(server.URL + "/list")
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	if len(result) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(result))
	}
}

func TestFlushDatabase(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Set a key
	http.Post(server.URL+"/set", "application/json", bytes.NewBuffer([]byte(`{"key":"flushKey", "value":"flushValue"}`)))

	// Flush the database
	resp, err := http.Post(server.URL+"/flush", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Check that the database is empty
	resp, err = http.Get(server.URL + "/list")
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	if len(result) != 0 {
		t.Errorf("Expected an empty database after flush, got %d keys", len(result))
	}
}

func TestCompactDatabase(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Set a key
	http.Post(server.URL+"/set", "application/json", bytes.NewBuffer([]byte(`{"key":"compactKey", "value":"compactValue"}`)))

	// Trigger compaction
	resp, err := http.Post(server.URL+"/compact", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Check if the key is still accessible after compaction
	resp, _ = http.Get(server.URL + "/get?key=compactKey")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected key to still exist after compaction, got %d", resp.StatusCode)
	}
}
