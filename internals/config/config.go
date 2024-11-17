package config

import (
	"fmt"
	"log"
	"loshon-api/internals/validator"
	"os"

	"github.com/spf13/viper"
)

type AppConfig struct {
	ClerkPublishableKey string `mapstructure:"CLERK_PUBLISHABLE_KEY" validate:"required"`
	ClerkSecretKey      string `mapstructure:"CLERK_SECRET_KEY" validate:"required"`
	PostgresUrl         string `mapstructure:"POSTGRES_URL" validate:"required"`
	AngoliaAppID        string `mapstructure:"ANGOLIA_APP_ID" validate:"required"`
	AngoliaAPIKey       string `mapstructure:"ANGOLIA_API_KEY" validate:"required"`
	Port                string `mapstructure:"PORT" validate:"required"`
}

func loadEnv(env string) (*AppConfig, error) {
	v := validator.NewValidator()
	config := AppConfig{}

	// auto load to env
	viper.AddConfigPath(".")
	viper.SetConfigName(fmt.Sprintf(".env.%s", env))
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to load config %v", err)
	}

	viper.Unmarshal(&config)
	if err := v.ValidateStruct(config); err != nil {
		log.Fatalf("failed to load config %v", err)
	}

	// if struct validation fail (missing required fields), try to load from file
	return &config, nil
}

func LoadConfig() (*AppConfig, error) {
	var env string
	if e, ok := os.LookupEnv("ENV"); !ok {
		env = "development"
	} else {
		env = e
	}

	return loadEnv(env)
}
