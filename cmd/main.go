package main

import (
	"log/slog"
	"os"

	"github.com/dafuqqqyunglean/avito_tech/app"
)

func main() {
	err := app.New().Run()
	if err != nil {
		slog.Warn("error occured", "error", err)
		os.Exit(1)
	}
}
