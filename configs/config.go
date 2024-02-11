package configs

import (
	"os"

	"github.com/protengplus/proteng-user-mgmt/internal/logger"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

var Config config

type config struct {
	Env      string `envconfig:"ENV" default:"dev"`
	HttpPort string `envconfig:"HTTP_PORT" default:"8080"`

	AccessTokenPrivateKey string `envconfig:"ACCESS_TOKEN_PRIVATE_KEY" default:""`
	AccessTokenExpiredIn  string `envconfig:"ACCESS_TOKEN_EXPIRED_IN" default:"1h"`
	AccessTokenMaxAge     int    `envconfig:"ACCESS_TOKEN_MAXAGE" default:"3600"`

	MongoUri string `envconfig:"MONGO_URI" default:""`
	MongoDb  string `envconfig:"MONGO_DB" default:"proteng-dev"`
}

// LoadConfig loads the config from the .env file if the ENV variable is set
func AutomaticLoadEnv() {
	if env, ok := os.LookupEnv("ENV"); ok {
		if err := LoadEnvFromPath(".env." + env); err != nil {
			logger.Fatalf("Error loading .env file from file: %v", err)
		} else {
			logger.Infof("Running in environment: %s", env)
		}
	}

	err := envconfig.Process("", &Config)
	if err != nil {
		logger.Fatalf("Error unmarshalling env vars: %v", err)
	}
}

func LoadEnvFromPath(path string) error {
	return godotenv.Load(path)
}
