package config

import (
	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	ServerPort   string
	OdooURL      string
	OdooDB       string
	OdooUsername string
	OdooPassword string
}

// Load reads configuration from a file and environment variables.
func Load() (*Config, error) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // or json, toml
	viper.AddConfigPath(".")      // look for config in the working directory
	viper.AutomaticEnv()          // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		// If the config file is not found, we can proceed
		// as long as environment variables are set.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Set default values
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("ODOO_URL", "https://maj.oneerp.app")
	viper.SetDefault("ODOO_DB", "maj_cn")
	viper.SetDefault("ODOO_USERNAME", "system")
	viper.SetDefault("ODOO_PASSWORD", "admindps.1.2.3@#@")


	cfg := &Config{
		ServerPort:   viper.GetString("SERVER_PORT"),
		OdooURL:      viper.GetString("ODOO_URL"),
		OdooDB:       viper.GetString("ODOO_DB"),
		OdooUsername: viper.GetString("ODOO_USERNAME"),
		OdooPassword: viper.GetString("ODOO_PASSWORD"),
	}

	return cfg, nil
}
