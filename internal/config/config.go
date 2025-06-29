package config

import (
	"fmt"
	"net"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	Development bool `mapstructure:"development"`

	PostgresConnectionString string `mapstructure:"postgres_connection_string"`

	ServerAddress string `mapstructure:"server_address"`
}

func setupConfigFile() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
}

func setupConfigEnv() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GODIN")
	viper.BindEnv("development", "DEVELOPMENT")
	viper.BindEnv("postgres_connection_string", "POSTGRES_CONNECTION_STRING")
}

func setupConfigDefaults() {
	viper.SetDefault("development", false)
	viper.SetDefault("postgres_connection_string", "")
	viper.SetDefault("server_address", ":60000")

}

func New() (*Config, error) {
	var config Config

	if err := gotenv.Load(); err != nil {
		panic(err)
	}

	setupConfigFile()
	setupConfigEnv()
	setupConfigDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.PostgresConnectionString == "" {
		return fmt.Errorf("postgres_connection_string is required")
	}

	if c.ServerAddress == "" {
		return fmt.Errorf("server_address is required")
	}

	if _, _, err := net.SplitHostPort(c.ServerAddress); err != nil {
		return fmt.Errorf("invalid server_address: %w", err)
	}

	return nil
}
