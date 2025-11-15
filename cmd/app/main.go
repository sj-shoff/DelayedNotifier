// cmd/app/main.go
package main

import (
	"delayed-notifier/internal/app"
	"delayed-notifier/internal/config"

	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.Init()

	cfg, err := config.MustLoad()
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to load config")
	}

	a := app.NewApp(cfg)
	a.Run()
}
