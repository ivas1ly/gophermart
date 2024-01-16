package main

import (
	"github.com/ivas1ly/gophermart/internal/app"
	"github.com/ivas1ly/gophermart/internal/config"
)

func main() {
	cfg := config.New()

	app.Run(cfg)
}
