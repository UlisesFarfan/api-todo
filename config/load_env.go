package config

import (
	"fmt"
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

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.AutomaticEnv()
	viper.SetConfigType("env")
	viper.SetConfigFile(".env")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	new_config := Config{
		DbUrl:          viper.GetString("MONGO_URL"),
		DbName:         viper.GetString("DATABASE"),
		ServerPort:     viper.GetString("PORT"),
		TokenSecret:    viper.GetString("TOKEN_SECRET"),
		TokenExpiresIn: viper.GetDuration("TOKEN_EXPIRED_IN"),
		TokenMaxAge:    viper.GetInt("TOKEN_MAXAGE"),
	}

	fmt.Println(new_config)

	return new_config, nil
}
