package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listen   string     `yaml:"listen"`
	Upstream string     `yaml:"upstream"`
	Chaos    FileConfig `yaml:"chaos"`

	UpstreamURL *url.URL `yaml:"-"`
}

type FileConfig struct {
	ErrorRate   float64 `yaml:"error_rate"`
	ErrorCode   int     `yaml:"error_code"`
	DropRate    float64 `yaml:"drop_rate"`
	Latency     string  `yaml:"latency"`
	LatencyMin  string  `yaml:"latency_min"`
	LatencyMax  string  `yaml:"latency_max"`
	CorruptRate float64 `yaml:"corrupt_rate"`
}

type Latencies struct {
	Latency    time.Duration
	LatencyMin time.Duration
	LatencyMax time.Duration
}

func Load() (*Config, error) {
	configFile := "config.yaml"

	var cfg Config

	// Read from config file
	_, err := os.Stat(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to find config file: %w", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %w", err)
	}

	if cfg.Upstream == "" {
		return nil, fmt.Errorf("upstream URL is required")
	}
	if cfg.Listen == "" {
		slog.Warn("listening port not defined, using default :8080")
		cfg.Listen = ":8080"
	}

	upstreamURL, err := url.Parse((cfg.Upstream))
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL: %w", err)
	}
	cfg.UpstreamURL = upstreamURL

	return &cfg, nil
}

func (cfg *Config) ParseDurations() (Latencies, error) {
	var lat Latencies

	if cfg.Chaos.Latency != "" {
		d, err := time.ParseDuration(cfg.Chaos.Latency)
		if err != nil {
			return Latencies{}, fmt.Errorf("invalid latency: %w", err)
		}
		lat.Latency = d
	} else {
		latencyMin, err := time.ParseDuration(cfg.Chaos.LatencyMin)
		if err != nil {
			return Latencies{}, fmt.Errorf("invalid latency: %w", err)
		}
		latencyMax, err := time.ParseDuration(cfg.Chaos.LatencyMax)
		if err != nil {
			return Latencies{}, fmt.Errorf("invalid latency: %w", err)
		}
		lat.LatencyMin = latencyMin
		lat.LatencyMax = latencyMax
	}

	return lat, nil
}

func (cfg *Config) PrintConfiguration() {
	fmt.Println("Chaos configuration")
	fmt.Printf("- Error rate: %v%%\n", cfg.Chaos.ErrorRate)
	fmt.Printf("- Error code: %v\n", cfg.Chaos.ErrorCode)
	fmt.Printf("- Drop rate: %v%%\n", cfg.Chaos.DropRate)

	if cfg.Chaos.Latency != "" {
		fmt.Printf("- Fixed latency: %v\n", cfg.Chaos.Latency)
	} else {
		fmt.Printf("- Minimum latency: %v\n", cfg.Chaos.LatencyMin)
		fmt.Printf("- Maximum latency: %v\n", cfg.Chaos.LatencyMax)

	}

	fmt.Printf("- Corrupt rate: %v%%\n", cfg.Chaos.CorruptRate)
}
