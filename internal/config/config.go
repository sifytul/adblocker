package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Upstream UpstreamConfig `yaml:"upstream"`
	Blocklist BlocklistConfig `yaml:"blocklist"`
	Logging LoggingConfig `yaml:"logging"`
}

type ServerConfig struct {
	ListenAddress string `yaml:"listen_address"`
	Protocol string `yaml:"protocol"`
}

type UpstreamConfig struct {
	Servers []string `yaml:"servers"`
	Timeout int
}

type BlocklistConfig struct {
	Sources []string `yaml:"sources"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	LogQueries bool `yaml:"log_queries"`
	LogBlocked bool `yaml:"log_blocked"`
	OutputFile string `yaml:"output_file"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			ListenAddress: "0.0.0.0:53",
			Protocol: "udp",
		},
		Upstream: UpstreamConfig{
			Servers: []string{"8.8.8.8:53", "1.1.1.1:53"},
			Timeout: 3,
		},
		Blocklist: BlocklistConfig{
			Sources: []string{"blocklists/test-blocklist.txt"},
		},
		Logging: LoggingConfig{
			Level: "info",
			LogQueries: true,
			LogBlocked: true,
			OutputFile: "",
		},
	}
}

func LoadFromFile(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configurateion: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.Server.ListenAddress == "" {
        return fmt.Errorf("server.listen_address cannot be empty")
    }
    
    // Check protocol
    if c.Server.Protocol != "udp" && c.Server.Protocol != "tcp" {
        return fmt.Errorf("server.protocol must be 'udp' or 'tcp'")
    }
    
    // Check upstream servers
    if len(c.Upstream.Servers) == 0 {
        return fmt.Errorf("at least one upstream server required")
    }
    
    // Check timeout
    if c.Upstream.Timeout <= 0 {
        return fmt.Errorf("upstream.timeout must be positive")
    }
    
    // Check blocklist sources
    if len(c.Blocklist.Sources) == 0 {
        return fmt.Errorf("at least one blocklist source required")
    }
    
	validLevels := map[string]bool{
		"debug": true,
		"warn": true,
		"info": true,
		"error": true,
	}

	if !validLevels[c.Logging.Level] {

		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	return nil
}