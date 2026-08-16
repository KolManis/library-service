package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL  string
	KafkaBrokers []string
	Port         string
}

func Load() (Config, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return Config{}, errors.New("DATABASE_URL не задан")
	}
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 1 && brokers[0] == "" {
		brokers = []string{"localhost:9092"}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{DatabaseURL: dsn, KafkaBrokers: brokers, Port: port}, nil
}
