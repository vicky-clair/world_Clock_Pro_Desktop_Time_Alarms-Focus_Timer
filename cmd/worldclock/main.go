package main

import (
	"log"
	"os"

	"worldclock/pkg/storage"
	"worldclock/pkg/ui"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/app/singleinstance"
	vui "github.com/vivaui/viva/ui"
)

func main() {
	store := storage.NewStore("")
	state := ui.NewAppState(store)

	// Disallow running multiple instances: if another copy is already running,
	// notify the primary instance to show/focus its window and exit cleanly.
	claim, primary, err := singleinstance.SingleInstance("com.worldclockpro.desktop.v1", os.Args[1:], func(args []string) {
		state.RequestShowWindow()
	})
	if err == nil && !primary {
		return
	}
	defer claim.Close()

	var canvas vui.Canvas
	cfg := ui.NewAppConfig(state)

	if err := app.RunApplication(func(application *app.Application) error {
		state.Application = application
		win, err := application.OpenWindow(cfg, &canvas)
		if err != nil {
			return err
		}
		state.Window = win

		application.OnReopen(func() {
			win.Show()
		})

		_, _ = application.AddStatusItem(app.StatusItemConfig{
			Tooltip: "World Clock Pro: Desktop Time, Alarms & Focus Timer",
			Icon:    ui.BuildStatusIcon(),
			OnClick: func(app.StatusClick) {
				win.Show()
			},
			Menu: func() []app.MenuItem {
				return []app.MenuItem{
					{
						Title: state.T("tray.show"),
						OnSelect: func() {
							win.Show()
						},
					},
					{
						Separator: true,
					},
					{
						Title: state.T("tray.exit"),
						OnSelect: func() {
							state.ForceQuit = true
							state.Persist()
							win.Close()
							application.Quit()
						},
					},
				}
			},
		})
		return nil
	}); err != nil {
		log.Fatalf("failed to run world clock application: %v", err)
	}
}
