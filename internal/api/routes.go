package api

import (
	"encoding/json"
	"net/http"
)

// handleSet handles setting a key
func (r *Router) handleSet(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Invalid Method"})
		return
	}

	var requestData struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	if requestData.Key == "" || requestData.Value == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Missing key or value"})
		return
	}

	if err := r.store.Set(requestData.Key, requestData.Value); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to set value"})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "Key set successfully"})
}

// handleGet retrieves a key's value
func (r *Router) handleGet(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Invalid Method"})
		return
	}

	key := req.URL.Query().Get("key")
	if key == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Missing key parameter"})
		return
	}

	value, err := r.store.Get(key)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Key not found"})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"key": key, "value": value})
}

// handleDelete removes a key
func (r *Router) handleDelete(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Invalid Method"})
		return
	}

	key := req.URL.Query().Get("key")
	if key == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Missing key parameter"})
		return
	}

	if err := r.store.Delete(key); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete key"})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "Key deleted successfully"})
}

// handleList returns all key-value pairs
func (r *Router) handleList(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Invalid Method"})
		return
	}

	data := r.store.List()
	jsonResponse(w, http.StatusOK, data)
}

// handleFlush clears all data
func (r *Router) handleFlush(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Invalid Method"})
		return
	}

	r.store.Flush()
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Database flushed"})
}

// handleCompact triggers data compaction
func (r *Router) handleCompact(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Invalid Method"})
		return
	}

	if err := r.store.CompactFlushedData(); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Compaction failed"})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "Compaction completed"})
}

// handleGetMemoryUsage returns the memory usage of the store
func (r *Router) handleGetMemoryUsage(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Invalid Method"})
		return
	}

	mem := r.store.MemoryUsage()
	jsonResponse(w, http.StatusOK, map[string]int{"memory": mem})
}

// handleGetKeyCount returns the number of keys in the store
func (r *Router) handleGetKeyCount(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Invalid Method"})
		return
	}

	count := r.store.KeyCount()
	jsonResponse(w, http.StatusOK, map[string]int{"count": count})
}

// handleBatchSet handles setting multiple keys in a batch
func (r *Router) handleBatchSet(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Only POST allowed"})
		return
	}

	var requestData []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
		return
	}

	count := 0
	for _, item := range requestData {
		if err := r.store.Set(item.Key, item.Value); err != nil {
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to set value"})
			return
		}
		count++
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":  "Keys set successfully",
		"keys_set": count,
	})
}

// handleBatchDelete handles deleting multiple keys in a batch
func (r *Router) handleBatchDelete(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Invalid Method"})
		return
	}

	var requestData []string
	if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
		return
	}

	count := 0
	for _, key := range requestData {
		if err := r.store.Delete(key); err != nil {
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete key"})
			return
		}
		count++
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":      "Keys deleted successfully",
		"keys_deleted": count,
	})
}
