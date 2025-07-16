package config

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	Development bool `mapstructure:"development"`

	PostgresConnectionString string `mapstructure:"postgres_connection_string"`

	ServerAddress string `mapstructure:"server_address"`
}

var configVars = map[string]any{
	"development":                false,
	"postgres_connection_string": "",
	"server_address":             ":60000",
}

func setupConfigFile() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
}

func bindToEnv(input string) error {
	return viper.BindEnv(input, strings.ToUpper(input))
}

func setupConfigEnv() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GODIN")

	for envVar, _ := range configVars {
		if err := bindToEnv(envVar); err != nil {
			panic(err)
		}
	}
}

func setupConfigDefaults() {
	for key, value := range configVars {
		viper.SetDefault(key, value)
	}
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
		var configFileNotFoundError viper.ConfigFileNotFoundError

		if !errors.As(err, &configFileNotFoundError) {
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
