// cmd/app/main.go
package main

import (
	"delayed-notifier/internal/app"
	"delayed-notifier/internal/config"
	"os"

	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.Init()

	cfg, err := config.MustLoad()
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to load config")
	}

	application, err := app.NewApp(cfg)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to create application")
	}

	if err := application.Run(); err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Application failed")
	}

	zlog.Logger.Info().Msg("Application exited successfully")
	os.Exit(0)
}
