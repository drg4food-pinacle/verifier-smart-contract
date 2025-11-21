package config

import (
	"fmt"
	"go-contracts/internal/types"
	"go-contracts/internal/validator"

	"github.com/spf13/viper"
)

type Config struct {
	*types.Config
}

func NewConfig() *Config {
	return &Config{
		&types.Config{
			LoggerMode: "development",
		},
	}
}

// LoadConfig loads the configuration from the .env file and validates it
func (c *Config) LoadConfig() error {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Load config from .env file
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Failed to read config file: %w", err)
	}

	// Set default values
	if err := viper.Unmarshal(&c.Config); err != nil {
		return fmt.Errorf("Failed to unmarshal config: %w", err)
	}

	// Validator
	if err := validator.ValidateStruct(c); err != nil {
		return fmt.Errorf("Config Validation Struct failed: %w", err)
	}

	return nil
}
