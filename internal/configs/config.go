// Package configs provides a struct for managing .env variables
package configs

import (
	"fmt"
	"log"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	*DBConfig
	*TgBotConfig
	*ElasticConfig
}

type TgBotConfig struct {
	Token string `env:"BOT_TOKEN"`
}

type DBConfig struct {
	DSN string `env:"DSN"`
}

type ElasticConfig struct {
	// URL is optional: when empty the bot falls back to Postgres airport search.
	URL string `env:"ELASTIC_URL" optional:"true"`
}

func LoadConfig() *Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("godotenv error: %s", err.Error())
	}

	conf := &Config{}
	errors := ParseConfig(conf)

	if len(errors) > 0 {
		printErrors(errors)
	}

	return conf
}

func printErrors(errs []error) {
	var msg strings.Builder
	msg.WriteString("config errors:\n")
	for _, err := range errs {
		fmt.Fprintf(&msg, "  - %s\n", err.Error())
	}
	log.Fatal(msg.String())
}
