package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWatcher_ValidFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte("test: data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	w, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	defer w.Close()

	if w.configPath != configPath {
		t.Errorf("Expected configPath to be '%s', got '%s'", configPath, w.configPath)
	}
	if w.reloadChan == nil {
		t.Error("Expected reloadChan to be initialized")
	}
	if w.watcher == nil {
		t.Error("Expected watcher to be initialized")
	}
}

func TestNewWatcher_InvalidFile(t *testing.T) {
	configPath := "/nonexistent/path/config.yaml"

	_, err := NewWatcher(configPath)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestWatcher_FileModification(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte("initial: data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	w, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	defer w.Close()

	// Start watching
	w.Start()

	// Start listening for signals in a goroutine before modifying the file
	signalReceived := make(chan bool, 1)
	go func() {
		select {
		case <-w.ReloadChan():
			signalReceived <- true
		case <-time.After(2 * time.Second):
			signalReceived <- false
		}
	}()

	// Give watcher time to start and begin listening
	time.Sleep(100 * time.Millisecond)

	// Modify the file
	err = os.WriteFile(configPath, []byte("modified: data"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Wait for result
	if received := <-signalReceived; !received {
		t.Error("Expected to receive reload signal, but timed out")
	}
}

func TestWatcher_MultipleModifications(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte("initial: data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	w, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	defer w.Close()

	w.Start()

	time.Sleep(100 * time.Millisecond) // Give watcher time to start

	// Modify file multiple times with different content
	modifications := []string{"modified1", "modified2", "modified3"}
	receivedSignals := 0

	for i, content := range modifications {
		// Modify the file
		err = os.WriteFile(configPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		// Try to receive signal with timeout
		select {
		case <-w.ReloadChan():
			receivedSignals++
		case <-time.After(1 * time.Second):
			t.Logf("Modification %d did not trigger signal within timeout", i+1)
		}

		// Small delay between modifications
		time.Sleep(100 * time.Millisecond)
	}

	if receivedSignals == 0 {
		t.Error("Expected to receive at least one reload signal")
	}
}

func TestWatcher_ReloadChan(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte("test: data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	w, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	defer w.Close()

	ch := w.ReloadChan()
	if ch == nil {
		t.Error("Expected ReloadChan to return a channel")
	}

	// Verify the channel is empty initially (non-blocking check)
	select {
	case <-ch:
		t.Error("Expected channel to be empty initially")
	case <-time.After(100 * time.Millisecond):
		// Success - channel is empty
	}
}

func TestWatcher_Close(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte("test: data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	w, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	w.Start()

	// Close the watcher
	err = w.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got: %v", err)
	}
}

func TestWatcher_IgnoreOtherFiles(t *testing.T) {
	// Create a temporary directory with multiple files
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	otherPath := filepath.Join(tmpDir, "other.yaml")

	err := os.WriteFile(configPath, []byte("config: data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	err = os.WriteFile(otherPath, []byte("other: data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create other file: %v", err)
	}

	w, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	defer w.Close()

	w.Start()

	// Start listening before modifying files
	signalChan := make(chan bool, 1)

	// Listen for signals from other file modification
	go func() {
		select {
		case <-w.ReloadChan():
			signalChan <- true
		case <-time.After(500 * time.Millisecond):
			signalChan <- false
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Modify the other file
	err = os.WriteFile(otherPath, []byte("other: modified"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify other file: %v", err)
	}

	// Should not receive reload signal for other file
	if received := <-signalChan; received {
		t.Error("Expected no reload signal for unrelated file modification")
	}

	// Now listen for config file modification
	go func() {
		select {
		case <-w.ReloadChan():
			signalChan <- true
		case <-time.After(2 * time.Second):
			signalChan <- false
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Modify the actual config file
	err = os.WriteFile(configPath, []byte("config: modified"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify config file: %v", err)
	}

	// Should receive reload signal for config file
	if received := <-signalChan; !received {
		t.Error("Expected reload signal for config file modification")
	}
}
