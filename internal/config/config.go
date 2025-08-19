package config

import (
        "github.com/robertguss/ai-news-agent-cli/internal/fetcher"
        "github.com/spf13/viper"
)

type Config struct {
        Sources []fetcher.Source `mapstructure:"sources"`
}

func Load() (*Config, error) {
        v := viper.New()
        v.SetConfigName("config")
        v.SetConfigType("yaml")
        v.AddConfigPath(".")
        v.AddConfigPath("./configs")
        
        if err := v.ReadInConfig(); err != nil {
                return nil, err
        }
        
        var cfg Config
        if err := v.Unmarshal(&cfg); err != nil {
                return nil, err
        }
        
        return &cfg, nil
}
