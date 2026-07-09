package config

import (
	"github.com/caarlos0/env/v10"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	OAuthTokenEncryptionKey string `env:"DATABASE_OAUTH_TOKEN_ENCRYPTION,required"`
}

var C Config

func init() {
	if err := env.Parse(&C); err != nil {
		panic(err)
	}
}
