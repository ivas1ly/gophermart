package main

import (
	"context"
	"log"

	"github.com/ivas1ly/gophermart/internal/app"
	"github.com/ivas1ly/gophermart/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.New()

	a, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Printf("can't init app: %s", err.Error())
		return
	}

	if err = a.Run(ctx); err != nil {
		log.Printf("application terminated with error: %s", err.Error())
	}
}
