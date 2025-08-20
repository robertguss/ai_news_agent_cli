package config

import (
	"os"
	"strconv"
	"time"

	"github.com/robertguss/ai-news-agent-cli/pkg/retry"
	"github.com/spf13/viper"
)

// Source represents a news source configuration with metadata for fetching and prioritization.
type Source struct {
	Name     string `mapstructure:"name"`
	URL      string `mapstructure:"url"`
	Type     string `mapstructure:"type"`
	Priority int    `mapstructure:"priority"`
}

// Config holds the complete application configuration including database settings,
// news sources, network timeouts, retry policies, and logging configuration.
type Config struct {
	DSN     string   `mapstructure:"dsn"`
	Sources []Source `mapstructure:"sources"`

	NetworkTimeout time.Duration `mapstructure:"network_timeout"`
	MaxRetries     int           `mapstructure:"max_retries"`
	BackoffBaseMs  int           `mapstructure:"backoff_base_ms"`
	BackoffMaxMs   int           `mapstructure:"backoff_max_ms"`
	DBBusyRetries  int           `mapstructure:"db_busy_retries"`
	LogFile        string        `mapstructure:"log_file"`
}

// Load loads the application configuration from the default config.yaml file.
// It searches for the config file in the current directory and ./configs directory.
func Load() (*Config, error) {
	return LoadFromPath("")
}

// LoadFromPath loads the application configuration from a specific file path.
// If configPath is empty, it uses the default search behavior.
func LoadFromPath(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.DSN == "" {
		cfg.DSN = "./ai-news.db"
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.NetworkTimeout == 0 {
		if timeoutStr := os.Getenv("NETWORK_TIMEOUT"); timeoutStr != "" {
			if timeout, err := time.ParseDuration(timeoutStr); err == nil {
				cfg.NetworkTimeout = timeout
			}
		}
		if cfg.NetworkTimeout == 0 {
			cfg.NetworkTimeout = 8 * time.Second
		}
	}

	if cfg.MaxRetries == 0 {
		if retriesStr := os.Getenv("MAX_RETRIES"); retriesStr != "" {
			if retries, err := strconv.Atoi(retriesStr); err == nil {
				cfg.MaxRetries = retries
			}
		}
		if cfg.MaxRetries == 0 {
			cfg.MaxRetries = 3
		}
	}

	if cfg.BackoffBaseMs == 0 {
		if baseStr := os.Getenv("BACKOFF_BASE_MS"); baseStr != "" {
			if base, err := strconv.Atoi(baseStr); err == nil {
				cfg.BackoffBaseMs = base
			}
		}
		if cfg.BackoffBaseMs == 0 {
			cfg.BackoffBaseMs = 250
		}
	}

	if cfg.BackoffMaxMs == 0 {
		if maxStr := os.Getenv("BACKOFF_MAX_MS"); maxStr != "" {
			if max, err := strconv.Atoi(maxStr); err == nil {
				cfg.BackoffMaxMs = max
			}
		}
		if cfg.BackoffMaxMs == 0 {
			cfg.BackoffMaxMs = 2000
		}
	}

	if cfg.DBBusyRetries == 0 {
		if retriesStr := os.Getenv("DB_BUSY_RETRIES"); retriesStr != "" {
			if retries, err := strconv.Atoi(retriesStr); err == nil {
				cfg.DBBusyRetries = retries
			}
		}
		if cfg.DBBusyRetries == 0 {
			cfg.DBBusyRetries = 3
		}
	}

	if cfg.LogFile == "" {
		if logFile := os.Getenv("LOG_FILE"); logFile != "" {
			cfg.LogFile = logFile
		} else {
			homeDir, _ := os.UserHomeDir()
			cfg.LogFile = homeDir + "/.ainews/agent.log"
		}
	}
}

func (c *Config) RetryConfig() retry.Config {
	return retry.Config{
		MaxRetries: c.MaxRetries,
		BaseDelay:  time.Duration(c.BackoffBaseMs) * time.Millisecond,
		MaxDelay:   time.Duration(c.BackoffMaxMs) * time.Millisecond,
		Multiplier: 2.0,
		MaxElapsed: c.NetworkTimeout * 2,
	}
}
