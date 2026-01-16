package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listen   string `yaml:"listen"`
	Upstream string `yaml:"upstream"`

	UpstreamURL *url.URL `yaml:"-"`
}

func Load() (*Config, error) {
	configFile := flag.String("config", "config.yaml", "Path to config file")
	listen := flag.String("listen", "", "Address to listen on (overrides config file)")
	upstream := flag.String("upstream", "", "Upstream URL to proxy to (override config file)")

	flag.Parse()

	cfg := &Config{
		Listen:   ":8080",
		Upstream: "",
	}

	// Read from config file
	if _, err := os.Stat(*configFile); err == nil {
		data, err := os.ReadFile(*configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file %w", err)
		}
	} else {
		fmt.Printf("No config file found at %s, using defaults\n", *configFile)
	}

	// Override if CLI flags provided
	if *listen != "" {
		cfg.Listen = *listen
	}
	if *upstream != "" {
		cfg.Upstream = *upstream
	}

	if cfg.Upstream == "" {
		return nil, fmt.Errorf("upstream URL is required (set in config file or --upstream flag)")
	}

	upstreamURL, err := url.Parse((cfg.Upstream))
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL: %w", err)
	}
	cfg.UpstreamURL = upstreamURL

	return cfg, nil
}
