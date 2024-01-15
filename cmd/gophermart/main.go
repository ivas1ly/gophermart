package main

import (
	"github.com/ivas1ly/gophermart/cmd/config"
	"github.com/ivas1ly/gophermart/internal/app"
)

func main() {
	cfg := config.New()

	app.Run(cfg)
}
