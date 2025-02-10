package engine

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const keyValueSeparator = " "

type Engine struct {
	data               map[string]string
	evictionQueue      []string // Keeps track of insertion order
	filePath           string
	flushPath          string
	memoryLimit        int
	currentMemoryUsage int
	mu                 sync.RWMutex
	fileMu             sync.Mutex
	saveChan           chan struct{}
	flushChan          chan struct{}
	shutdownChan       chan struct{} // For graceful shutdown
}

type EngineConfig struct {
	FilePath    string
	FlushPath   string
	MemoryLimit int
}

// NewEngine creates a new instance of Engine with the specified file path, flush path, and memory limit.
// It initializes the internal data structures and starts background workers for auto-saving and auto-flushing.
//
// Parameters:
//   - filePath: The path to the file where data will be stored.
//   - flushPath: The path to the file where data will be flushed.
//   - memoryLimit: The memory limit for the engine.
//
// Returns:
//   - A pointer to the new Engine instance.
//   - An error if initialization fails.
func NewEngine(filePath, flushPath string, memoryLimit int) (*Engine, error) {
	if memoryLimit <= 0 {
		return nil, errors.New("memory limit must be positive")
	}
	if filePath == "" || flushPath == "" {
		return nil, errors.New("file paths cannot be empty")
	}

	e := &Engine{
		data:         make(map[string]string),
		filePath:     filePath,
		flushPath:    flushPath,
		memoryLimit:  memoryLimit,
		saveChan:     make(chan struct{}, 1),
		flushChan:    make(chan struct{}, 1),
		shutdownChan: make(chan struct{}),
	}

	if err := e.Load(); err != nil {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	// Start background workers
	go e.autoSaveWorker()
	go e.autoFlushWorker()

	return e, nil
}

// Shutdown gracefully stops the background workers.
func (e *Engine) Shutdown() {
	close(e.shutdownChan)
}

// Set adds or updates a key-value pair and triggers async saving or flushing.
func (e *Engine) Set(key, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	oldSize := 0
	if oldVal, exists := e.data[key]; exists {
		oldSize = len(oldVal) + len(key)
	} else {
		e.evictionQueue = append(e.evictionQueue, key) // Track insertion order
	}

	newSize := len(value) + len(key)
	e.currentMemoryUsage = e.currentMemoryUsage - oldSize + newSize
	e.data[key] = value

	// If memory exceeds limit, trigger flush
	if e.currentMemoryUsage >= e.memoryLimit {
		select {
		case e.flushChan <- struct{}{}:
		default:
		}
	} else {
		// Otherwise, trigger async save
		select {
		case e.saveChan <- struct{}{}:
		default:
		}
	}

	return nil
}

// Get retrieves a value by key.
func (e *Engine) Get(key string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	value, ok := e.data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return value, nil
}

// Delete removes a key-value pair and triggers async saving.
func (e *Engine) Delete(key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if value, exists := e.data[key]; exists {
		e.currentMemoryUsage -= len(key) + len(value)
		delete(e.data, key)

		// Remove from eviction queue
		e.removeFromEvictionQueue(key)

		// Trigger async save
		select {
		case e.saveChan <- struct{}{}:
		default:
		}
	}
	return nil
}

// removeFromEvictionQueue removes a key from the eviction tracking queue.
func (e *Engine) removeFromEvictionQueue(key string) {
	for i, k := range e.evictionQueue {
		if k == key {
			e.evictionQueue = append(e.evictionQueue[:i], e.evictionQueue[i+1:]...)
			break
		}
	}
}

// Flush clears in-memory data and persists the change.
func (e *Engine) Flush() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.data = make(map[string]string)
	e.evictionQueue = []string{}
	e.currentMemoryUsage = 0

	// Trigger async save
	select {
	case e.saveChan <- struct{}{}:
	default:
	}
}

// SaveFile writes only the latest data to disk, avoiding duplicate keys.
func (e *Engine) SaveFile(data map[string]string) error {
	e.fileMu.Lock()
	defer e.fileMu.Unlock()

	file, err := os.OpenFile(e.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 64*1024) // 64 KB buffer
	for key, value := range data {
		if _, err := writer.WriteString(key + keyValueSeparator + value + "\n"); err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
	}
	return writer.Flush()
}

// autoSaveWorker periodically saves data when triggered.
func (e *Engine) autoSaveWorker() {
	for {
		select {
		case <-e.saveChan:
			time.Sleep(1 * time.Second) // Debounce multiple save requests

			e.mu.RLock()
			dataCopy := make(map[string]string, len(e.data))
			for k, v := range e.data {
				dataCopy[k] = v
			}
			e.mu.RUnlock()

			if err := e.SaveFile(dataCopy); err != nil {
				log.Printf("Error saving data: %v\n", err)
			}
		case <-e.shutdownChan:
			return
		}
	}
}

// AppendFlushedData appends flushed data to the flush file.
func (e *Engine) AppendFlushedData(data map[string]string) error {
	e.fileMu.Lock()
	defer e.fileMu.Unlock()

	file, err := os.OpenFile(e.flushPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open flush file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 64*1024) // 64 KB buffer
	for key, value := range data {
		if _, err := writer.WriteString(key + keyValueSeparator + value + "\n"); err != nil {
			return fmt.Errorf("failed to write flushed data: %w", err)
		}
	}
	return writer.Flush()
}

// autoFlushWorker removes just enough old data when memory usage exceeds the limit.
func (e *Engine) autoFlushWorker() {
	for {
		select {
		case <-e.flushChan:
			e.mu.Lock()

			if e.currentMemoryUsage < e.memoryLimit {
				e.mu.Unlock()
				continue
			}

			bytesToFree := e.currentMemoryUsage - e.memoryLimit
			evictedData := make(map[string]string)
			freedBytes := 0

			log.Printf("Memory limit exceeded! Flushing oldest keys to flushed.db... currentMemoryUsage: %d, memoryLimit: %d\n", e.currentMemoryUsage, e.memoryLimit)

			for len(e.evictionQueue) > 0 && freedBytes < bytesToFree {
				key := e.evictionQueue[0] // Remove oldest key first
				e.evictionQueue = e.evictionQueue[1:]

				if val, exists := e.data[key]; exists {
					evictedData[key] = val
					freedBytes += len(key) + len(val)
					delete(e.data, key)
				}
			}

			e.currentMemoryUsage -= freedBytes
			e.mu.Unlock()

			log.Printf("Flushed keys: %d, Freed bytes: %d\n", len(evictedData), freedBytes)

			// Save flushed data separately
			if len(evictedData) > 0 {
				if err := e.AppendFlushedData(evictedData); err != nil {
					log.Printf("Error saving flushed data: %v\n", err)
				}
			}
		case <-e.shutdownChan:
			return
		}
	}
}

// Load loads data from disk into memory.
func (e *Engine) Load() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Reset in-memory data structures
	e.data = make(map[string]string)
	e.evictionQueue = []string{}

	// Load primary data from data.db
	dataFromFile, err := e.loadFromFile(e.filePath)
	if err != nil {
		return fmt.Errorf("failed to load data from file: %w", err)
	}

	// Load evicted (flushed) data from flushed.db
	flushedData, err := e.loadFromFile(e.flushPath)
	if err != nil {
		return fmt.Errorf("failed to load flushed data: %w", err)
	}

	// Merge flushed data into main memory (flushed keys take precedence)
	for key, value := range dataFromFile {
		e.data[key] = value
		e.evictionQueue = append(e.evictionQueue, key)
	}
	for key, value := range flushedData {
		e.data[key] = value // Overwrite if key was in `data.db`
		e.evictionQueue = append(e.evictionQueue, key)
		e.currentMemoryUsage += len(key) + len(value)
	}

	log.Println("Load complete: Memory store restored from disk.")
	return nil
}

// loadFromFile loads key-value pairs from a given file.
func (e *Engine) loadFromFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil // Return an empty map if file doesn't exist
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, keyValueSeparator, 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		data[key] = value
	}

	return data, scanner.Err()
}

// CompactFlushedData ensures flushed data is merged into data.db and safely deleted.
func (e *Engine) CompactFlushedData() error {
	log.Println("Starting compaction of flushed data...")

	// Step 1: Pause auto-save to avoid race conditions
	e.mu.Lock()
	saveWorker := e.saveChan // Capture save worker state
	e.saveChan = nil         // Disable save worker
	e.mu.Unlock()

	// Step 2: Ensure any pending flush completes before compaction
	err := e.forceFlush()
	if err != nil {
		return fmt.Errorf("force flush failed: %w", err)
	}
	log.Println("Forced flush completed")

	// Step 3: Load flushed data **without locking**
	flushedData, err := e.loadFromFile(e.flushPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load flushed data: %w", err)
	}
	log.Println("Flushed data loaded")

	// Step 4: Load existing data from data.db **without locking**
	existingData, err := e.loadFromFile(e.filePath)
	if err != nil {
		return fmt.Errorf("failed to load existing data: %w", err)
	}
	log.Println("Existing data loaded")

	// Step 5: Merge flushed data into existing data
	for key, value := range flushedData {
		existingData[key] = value
	}
	log.Println("Data merged")

	// Step 6: Save merged data to main data file
	if err := e.SaveFile(existingData); err != nil {
		return fmt.Errorf("failed to save merged data: %w", err)
	}
	log.Println("Merged data saved")

	// Step 7: Update memory usage accurately
	e.mu.Lock()
	e.currentMemoryUsage = 0
	for key, value := range existingData {
		e.currentMemoryUsage += len(key) + len(value)
	}
	e.saveChan = saveWorker // Restore save worker after compaction
	e.mu.Unlock()

	log.Println("Memory usage updated")

	// Step 8: Remove flushed.db after unlocking
	if err := os.Remove(e.flushPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove flushed file: %w", err)
	}
	log.Println("Flushed data removed")

	log.Println("Compaction completed successfully.")
	return nil
}

// Save persists the current in-memory data to disk.
func (e *Engine) Save() error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.SaveFile(e.data)
}

// PrintMemoryUsage prints memory statistics.
func (e *Engine) PrintMemoryUsage() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	log.Printf("HeapAlloc: %d KB, TotalAlloc: %d KB, Sys: %d KB\n", memStats.HeapAlloc/1024, memStats.TotalAlloc/1024, memStats.Sys/1024)
	log.Printf("Current Store Memory Usage: %d bytes\n", e.currentMemoryUsage)
}

// Testing Helpers
func (e *Engine) MemoryUsage() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.currentMemoryUsage
}

func (e *Engine) DataSize() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.data)
}

// List returns a copy of the in-memory data.
func (e *Engine) List() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	copy := make(map[string]string)
	for k, v := range e.data {
		copy[k] = v
	}
	return copy
}

// forceFlush ensures that all in-memory data exceeding memoryLimit is saved before compaction
func (e *Engine) forceFlush() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.currentMemoryUsage < e.memoryLimit {
		return nil // No flush needed
	}

	// Move current data to flushed.db
	evictedData := make(map[string]string)
	for key, value := range e.data {
		evictedData[key] = value
		delete(e.data, key)
	}

	// Reset memory usage
	e.currentMemoryUsage = 0

	// Save flushed data to flushed.db
	return e.AppendFlushedData(evictedData)
}

func (e *Engine) KeyCount() int {
	return len(e.data)
}

func (e *Engine) GetMemoryLimit() int {
	return e.memoryLimit
}

func (e *Engine) SetMemoryLimit(limit int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.memoryLimit = limit
}

func (e *Engine) LoadConfig(config EngineConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.filePath = config.FilePath
	e.flushPath = config.FlushPath
	e.memoryLimit = config.MemoryLimit

	return e.Load()
}
