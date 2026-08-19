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

	EmailFrom string `envconfig:"EMAIL_FROM" default:"proteng.plus@gmail.com"`
	SMTPHost  string `envconfig:"SMTP_HOST"`
	SMTPPass  string `envconfig:"SMTP_PASS"`
	SMTPPort  int    `envconfig:"SMTP_PORT"`
	SMTPUser  string `envconfig:"SMTP_USER"`
	Origin    string `envconfig:"ORIGIN"`

	RabbitMqUrl string `envconfig:"RABBITMQ_URL"`
	JobQueue    string `envconfig:"JOB_QUEUE"`
}

// LoadConfig loads the config from the .env file if the ENV variable is set
func AutomaticLoadEnv() {
	if env, ok := os.LookupEnv("ENV"); ok {
		if err := LoadEnvFromPath(".env." + env); err != nil {
			logger.Infof("No .env.%s file found, using environment variables as-is: %v", env, err)
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
