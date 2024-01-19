package main

import (
	"log"

	"github.com/ivas1ly/gophermart/internal/app"
	"github.com/ivas1ly/gophermart/internal/config"
)

func main() {
	cfg := config.New()

	if err := app.Run(cfg); err != nil {
		log.Printf("application terminated with error: %v", err)
	}
}
