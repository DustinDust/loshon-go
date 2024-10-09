package config

import (
	"github.com/spf13/viper"
)

type EnvConfig struct {
	ClerkPublishableKey string `mapstructure:"CLERK_PUBLISHABLE_KEY"`
	ClerkSecretKey      string `mapstructure:"CLERK_SECRET_KEY"`
	PostgresUrl         string `mapstructure:"POSTGRES_URL"`
	Environment         string `mapstructure:"ENV"`
	Addr                string `mapstructure:"ADDR"`
}

func loadEnv(path string) (*EnvConfig, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	config := &EnvConfig{}
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	return config, err
}

func LoadConfig() (*EnvConfig, error) {
	return loadEnv(".")
}
