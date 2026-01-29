package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func Init() {
	_ = godotenv.Load()
	viper.AutomaticEnv()
	viper.SetDefault("GIN_MODE", "debug")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DB_SSLMODE", "disable")
}
