package engine_test

import (
	"os"
	"testing"
	"time"

	"github.com/bendigiorgio/go-kv/internal/engine"
)

const TEST_FILE_PATH = "test_data.db"
const TEST_FLUSH_PATH = "test_flush.db"

// Helper function to create a fresh engine instance
func setupEngine(memoryLimit int) *engine.Engine {
	_ = os.Remove(TEST_FILE_PATH) // Ensure a fresh start
	engine, _ := engine.NewEngine(TEST_FILE_PATH, TEST_FLUSH_PATH, memoryLimit)
	return engine
}

func Test_SetAndGet(t *testing.T) {
	db := setupEngine(1024)

	err := db.Set("name", "Alice")
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	value, err := db.Get("name")
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}

	if value != "Alice" {
		t.Errorf("Expected 'Alice', got '%s'", value)
	}
}

func Test_GetNonExistentKey(t *testing.T) {
	db := setupEngine(1024)

	_, err := db.Get("unknown")
	if err == nil {
		t.Error("Expected error for non-existent key, got nil")
	}
}

func Test_DeleteKey(t *testing.T) {
	db := setupEngine(1024)

	_ = db.Set("name", "Alice")
	err := db.Delete("name")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	_, err = db.Get("name")
	if err == nil {
		t.Error("Expected error after deletion, got nil")
	}
}

func Test_ListKeys(t *testing.T) {
	db := setupEngine(1024)

	_ = db.Set("name", "Alice")
	_ = db.Set("age", "25")

	list := db.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 items, got %d", len(list))
	}

	if list["name"] != "Alice" || list["age"] != "25" {
		t.Errorf("List() returned incorrect data: %v", list)
	}
}

func Test_Flush(t *testing.T) {
	db := setupEngine(1024)

	_ = db.Set("name", "Alice")
	db.Flush()

	if len(db.List()) != 0 {
		t.Error("Flush() did not clear data")
	}
}

func Test_SaveAndLoad(t *testing.T) {
	db := setupEngine(1024)

	_ = db.Set("name", "Alice")
	_ = db.Set("city", "New York")

	// Save to disk
	err := db.Save()
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Create a new engine and load from file
	db2, _ := engine.NewEngine(TEST_FILE_PATH, TEST_FLUSH_PATH, 1024)

	value, err := db2.Get("name")
	if err != nil || value != "Alice" {
		t.Errorf("Load() did not retrieve 'name' correctly, got '%s', error: %v", value, err)
	}

	value, err = db2.Get("city")
	if err != nil || value != "New York" {
		t.Errorf("Load() did not retrieve 'city' correctly, got '%s', error: %v", value, err)
	}
}

func Test_MemoryLimitTriggersRollingFlush(t *testing.T) {
	// Set a small memory limit to force a rolling flush quickly
	db := setupEngine(50)

	// Insert multiple keys that will exceed memory
	_ = db.Set("k1", "value1")
	_ = db.Set("k2", "value2")
	_ = db.Set("k3", "value3")
	_ = db.Set("k4", "value4")

	// Wait for flush to complete
	time.Sleep(2 * time.Second)

	// Ensure at least some keys remain in memory
	if db.DataSize() == 0 {
		t.Error("Memory limit exceeded but all data was removed instead of rolling flush")
	}

	// Load a new engine and check data is still retrievable
	db2, _ := engine.NewEngine(TEST_FILE_PATH, TEST_FLUSH_PATH, 50)

	_, err := db2.Get("k1")
	_, err2 := db2.Get("k2")
	if err != nil && err2 != nil {
		t.Error("Rolling flush removed too much data; expected some keys to be recoverable")
	}
}

func Test_OverwriteExistingKey(t *testing.T) {
	db := setupEngine(1024)

	_ = db.Set("key", "oldValue")
	_ = db.Set("key", "newValue")

	value, _ := db.Get("key")
	if value != "newValue" {
		t.Errorf("Expected 'newValue', got '%s'", value)
	}
}

func Test_MemoryUsageTracking(t *testing.T) {
	db := setupEngine(1024)

	_ = db.Set("small", "test")
	initialMemory := db.MemoryUsage()

	_ = db.Set("big", "this is a much bigger value")
	if db.MemoryUsage() <= initialMemory {
		t.Error("Memory usage did not increase as expected")
	}

	_ = db.Delete("big")
	if db.MemoryUsage() != initialMemory {
		t.Error("Memory usage did not decrease after deletion")
	}
}

func Test_EnsureDataPersistsAcrossInstances(t *testing.T) {
	db := setupEngine(1024)

	_ = db.Set("persistentKey", "PersistentData")
	_ = db.Save()

	// Create a new engine instance to test persistence
	db2, _ := engine.NewEngine(TEST_FILE_PATH, TEST_FLUSH_PATH, 1024)

	value, err := db2.Get("persistentKey")
	if err != nil {
		t.Fatalf("Expected key to persist, but got error: %v", err)
	}
	if value != "PersistentData" {
		t.Errorf("Expected 'PersistentData', got '%s'", value)
	}
}

func Test_EnsureLRUFlushLogic(t *testing.T) {
	db := setupEngine(300)

	// Insert 5 keys that will exceed memory limit
	_ = db.Set("a", "dataA")
	_ = db.Set("b", "dataB")
	_ = db.Set("c", "dataC")
	_ = db.Set("d", "dataD")
	_ = db.Set("e", "dataE")

	// Wait for the flush to occur
	time.Sleep(2 * time.Second)

	// Ensure some keys were removed, but some remain
	if db.DataSize() == 0 {
		t.Error("All keys were flushed instead of maintaining recent keys")
	}

	// Reload from disk to check persistence
	db2, _ := engine.NewEngine(TEST_FILE_PATH, TEST_FLUSH_PATH, 100)

	// Some keys should still be available after the flush
	_, err := db2.Get("c")
	_, err2 := db2.Get("d")
	if err != nil && err2 != nil {
		t.Error("LRU logic removed too many keys; expected recent keys to be available")
	}
}

func Test_CompactFlushedData(t *testing.T) {
	db := setupEngine(50) // Set low memory limit to force flush

	// Step 1: Insert multiple keys to exceed memory limit and trigger a flush
	_ = db.Set("key1", "value1")
	_ = db.Set("key2", "value2")
	_ = db.Set("key3", "value3")
	_ = db.Set("key4", "value4")
	_ = db.Set("key5", "value5")
	_ = db.Set("key6", "value6")

	// Step 2: Wait for flush to complete
	time.Sleep(2 * time.Second)

	// Step 3: **Wait briefly to allow flush worker to complete**
	time.Sleep(100 * time.Millisecond)

	// Step 4: Ensure flushed.db is created
	if _, err := os.Stat(TEST_FLUSH_PATH); os.IsNotExist(err) {
		t.Fatalf("Flushed data file not found; expected a flush to occur")
	}

	// Step 5: Perform compaction
	err := db.CompactFlushedData()
	if err != nil {
		t.Fatalf("Compaction failed: %v", err)
	}

	// Step 6: Ensure flushed.db is deleted after compaction
	if _, err := os.Stat(TEST_FLUSH_PATH); !os.IsNotExist(err) {
		t.Fatalf("Flushed data file still exists after compaction")
	}

	// Step 7: Ensure all keys are still retrievable
	_, err1 := db.Get("key1")
	_, err2 := db.Get("key2")
	_, err3 := db.Get("key3")
	_, err4 := db.Get("key4")
	_, err5 := db.Get("key5")
	_, err6 := db.Get("key6")

	if err1 == nil || err2 == nil || err3 == nil || err4 == nil || err5 == nil || err6 == nil {
		t.Fatalf("Some keys are missing after compaction")
	}

	// Step 8: Ensure memory usage is updated correctly
	if db.MemoryUsage() == 0 {
		t.Fatalf("Memory usage should be updated after compaction")
	}
}
