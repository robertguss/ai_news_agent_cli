package config

import (
        "github.com/robertguss/ai-news-agent-cli/internal/fetcher"
        "github.com/spf13/viper"
)

type Config struct {
        DSN     string            `mapstructure:"dsn"`
        Sources []fetcher.Source `mapstructure:"sources"`
}

func Load() (*Config, error) {
        return LoadFromPath("")
}

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
        
        return &cfg, nil
}
