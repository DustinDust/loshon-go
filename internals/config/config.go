package config

import (
	"fmt"
	"loshon-api/internals/validator"

	"github.com/spf13/viper"
)

type AppConfig struct {
	ClerkPublishableKey string `mapstructure:"CLERK_PUBLISHABLE_KEY" validate:"required"`
	ClerkSecretKey      string `mapstructure:"CLERK_SECRET_KEY" validate:"required"`
	PostgresUrl         string `mapstructure:"POSTGRES_URL" validate:"required"`
	Environment         string `mapstructure:"ENV" validate:"required"`
	Addr                string `mapstructure:"ADDR" validate:"required"`
}

func (aconf *AppConfig) Merge(conf *AppConfig) {
	if aconf.ClerkPublishableKey == "" {
		aconf.ClerkPublishableKey = conf.ClerkPublishableKey
	}

	if aconf.ClerkSecretKey == "" {
		aconf.ClerkSecretKey = conf.ClerkSecretKey
	}

	if aconf.PostgresUrl == "" {
		aconf.PostgresUrl = conf.PostgresUrl
	}

	if aconf.Environment == "" {
		aconf.Environment = conf.Environment
	}

	if aconf.Addr == "" {
		aconf.Addr = conf.Addr
	}
}

func loadEnv(altPath string) (*AppConfig, error) {
	// auto load to env
	viper.AutomaticEnv()

	config := AppConfig{
		ClerkPublishableKey: viper.GetString("CLERK_PUBLISHABLE_KEY"),
		ClerkSecretKey:      viper.GetString("CLERK_SECRET_KEY"),
		PostgresUrl:         viper.GetString("POSTGRES_URL"),
		Environment:         viper.GetString("ENVIRONMENT"),
		Addr:                viper.GetString("ADDR"),
	}

	v := validator.NewValidator()
	err := v.ValidateStruct(config)

	if err != nil {
		viper.AddConfigPath(altPath)
		viper.SetConfigType("env")
		viper.SetConfigFile(".env")

		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("error loading alternative config: %v", err.Error())
		}

		conf := &AppConfig{}
		if err := viper.Unmarshal(&conf); err != nil {
			return nil, fmt.Errorf("error loading alternative config: %v", err.Error())
		}

		config.Merge(conf)
		if err := v.ValidateStruct(config); err != nil {
			return nil, fmt.Errorf("error loading alternative config: %v", err.Error())
		}
	}
	return &config, nil
}

func LoadConfig() (*AppConfig, error) {
	return loadEnv(".")
}
