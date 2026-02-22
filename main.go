package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	visible := true

	err := wails.Run(&options.App{
		Title:            "lay",
		Width:            820,
		Height:           580,
		Frameless:        true,
		AlwaysOnTop:      true,
		BackgroundColour: &options.RGBA{R: 18, G: 18, B: 24, A: 220},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			ProtectWindow()
			RegisterGlobalHotkey()

			// Listen for ⌘+Shift+L and toggle window visibility.
			go func() {
				for range hotkeyChannel {
					if visible {
						wailsRuntime.WindowHide(ctx)
					} else {
						wailsRuntime.WindowShow(ctx)
					}
					visible = !visible
				}
			}()
		},
		OnShutdown: func(ctx context.Context) {
			UnregisterGlobalHotkey()
		},
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			TitleBar:             mac.TitleBarHiddenInset(),
			Appearance:           mac.NSAppearanceNameDarkAqua,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
