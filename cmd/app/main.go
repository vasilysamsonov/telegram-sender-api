package main

import (
	"log"

	"telegram-sender-api/config"
	"telegram-sender-api/internal/app"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err = app.Run(cfg); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
