package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DbUrl  string `mapstructure:"MONGO_URL"`
	DbName string `mapstructure:"DATABASE"`

	ServerPort string `mapstructure:"PORT"`

	TokenSecret    string        `mapstructure:"TOKEN_SECRET"`
	TokenExpiresIn time.Duration `mapstructure:"TOKEN_EXPIRED_IN"`
	TokenMaxAge    int           `mapstructure:"TOKEN_MAXAGE"`
}

func LoadConfig() (config Config, err error) {
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetConfigType("env")
	viper.SetConfigFile(".env")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
