package main

import (
	"log"

	"worldclock/pkg/storage"
	"worldclock/pkg/ui"

	"github.com/vivaui/viva/app"
	vui "github.com/vivaui/viva/ui"
)

func main() {
	store := storage.NewStore("")
	state := ui.NewAppState(store)

	var canvas vui.Canvas
	cfg := ui.NewAppConfig(state)

	if err := app.RunCanvas(cfg, &canvas); err != nil {
		log.Fatalf("failed to run world clock application: %v", err)
	}
}
