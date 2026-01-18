package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `listen: ":9090"
upstream: "http://localhost:8080"
chaos:
  error_rate: 10.5
  error_code: 503
  drop_rate: 5.0
  latency: "100ms"
  corrupt_rate: 15.0
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Change to temp directory to test Load()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Listen != ":9090" {
		t.Errorf("Expected listen to be ':9090', got '%s'", cfg.Listen)
	}
	if cfg.Upstream != "http://localhost:8080" {
		t.Errorf("Expected upstream to be 'http://localhost:8080', got '%s'", cfg.Upstream)
	}
	if cfg.Chaos.ErrorRate != 10.5 {
		t.Errorf("Expected error_rate to be 10.5, got %v", cfg.Chaos.ErrorRate)
	}
	if cfg.Chaos.ErrorCode != 503 {
		t.Errorf("Expected error_code to be 503, got %v", cfg.Chaos.ErrorCode)
	}
	if cfg.UpstreamURL == nil {
		t.Error("Expected UpstreamURL to be parsed")
	}
	if cfg.UpstreamURL.Host != "localhost:8080" {
		t.Errorf("Expected upstream host to be 'localhost:8080', got '%s'", cfg.UpstreamURL.Host)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing config file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidContent := `listen: ":9090"
upstream: "http://localhost:8080
  invalid_yaml: [unclosed
`

	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	_, err = Load()
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestLoad_MissingUpstream(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `listen: ":9090"
chaos:
  error_rate: 10
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	_, err = Load()
	if err == nil {
		t.Error("Expected error for missing upstream, got nil")
	}
}

func TestLoad_InvalidUpstreamURL(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `listen: ":9090"
upstream: "://invalid-url"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	_, err = Load()
	if err == nil {
		t.Error("Expected error for invalid upstream URL, got nil")
	}
}

func TestLoad_DefaultListen(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `upstream: "http://localhost:8080"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Listen != ":8080" {
		t.Errorf("Expected default listen to be ':8080', got '%s'", cfg.Listen)
	}
}

func TestParseDurations_FixedLatency(t *testing.T) {
	cfg := &Config{
		Chaos: FileConfig{
			Latency: "250ms",
		},
	}

	latencies, err := cfg.ParseDurations()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if latencies.Latency != 250*time.Millisecond {
		t.Errorf("Expected latency to be 250ms, got %v", latencies.Latency)
	}
	if latencies.LatencyMin != 0 {
		t.Errorf("Expected LatencyMin to be 0, got %v", latencies.LatencyMin)
	}
	if latencies.LatencyMax != 0 {
		t.Errorf("Expected LatencyMax to be 0, got %v", latencies.LatencyMax)
	}
}

func TestParseDurations_RandomLatency(t *testing.T) {
	cfg := &Config{
		Chaos: FileConfig{
			LatencyMin: "100ms",
			LatencyMax: "500ms",
		},
	}

	latencies, err := cfg.ParseDurations()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if latencies.LatencyMin != 100*time.Millisecond {
		t.Errorf("Expected LatencyMin to be 100ms, got %v", latencies.LatencyMin)
	}
	if latencies.LatencyMax != 500*time.Millisecond {
		t.Errorf("Expected LatencyMax to be 500ms, got %v", latencies.LatencyMax)
	}
	if latencies.Latency != 0 {
		t.Errorf("Expected Latency to be 0, got %v", latencies.Latency)
	}
}

func TestParseDurations_InvalidFixedLatency(t *testing.T) {
	cfg := &Config{
		Chaos: FileConfig{
			Latency: "invalid",
		},
	}

	_, err := cfg.ParseDurations()
	if err == nil {
		t.Error("Expected error for invalid latency format, got nil")
	}
}

func TestParseDurations_InvalidMinLatency(t *testing.T) {
	cfg := &Config{
		Chaos: FileConfig{
			LatencyMin: "invalid",
			LatencyMax: "500ms",
		},
	}

	_, err := cfg.ParseDurations()
	if err == nil {
		t.Error("Expected error for invalid latency_min format, got nil")
	}
}

func TestParseDurations_InvalidMaxLatency(t *testing.T) {
	cfg := &Config{
		Chaos: FileConfig{
			LatencyMin: "100ms",
			LatencyMax: "invalid",
		},
	}

	_, err := cfg.ParseDurations()
	if err == nil {
		t.Error("Expected error for invalid latency_max format, got nil")
	}
}

func TestPrintConfiguration(t *testing.T) {
	cfg := &Config{
		Chaos: FileConfig{
			ErrorRate:   10.5,
			ErrorCode:   503,
			DropRate:    5.0,
			Latency:     "100ms",
			CorruptRate: 15.0,
		},
	}

	// This test just ensures PrintConfiguration doesn't panic
	// In a real scenario, you might want to capture stdout
	cfg.PrintConfiguration()
}

func TestPrintConfiguration_RandomLatency(t *testing.T) {
	cfg := &Config{
		Chaos: FileConfig{
			ErrorRate:   10.5,
			ErrorCode:   503,
			DropRate:    5.0,
			LatencyMin:  "100ms",
			LatencyMax:  "500ms",
			CorruptRate: 15.0,
		},
	}

	// This test just ensures PrintConfiguration doesn't panic with random latency
	cfg.PrintConfiguration()
}
