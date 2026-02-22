package main

import (
	"context"
	"embed"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "lay",
		Width:            520,
		Height:           360,
		MinWidth:         520,
		MinHeight:        360,
		Frameless:        true,
		AlwaysOnTop:      true,
		BackgroundColour: &options.RGBA{R: 18, G: 18, B: 24, A: 220},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			ProtectWindow()
			SetAccessoryPolicy()
			RegisterGlobalHotkey()
			go func() {
				time.Sleep(75 * time.Millisecond)
				positionWindowTopRight(ctx)
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
		Windows: &windows.Options{
			// Keep Windows 11 rounded corners + shadow in frameless mode.
			DisableFramelessWindowDecorations: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func positionWindowTopRight(ctx context.Context) {
	screens, err := runtime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return
	}

	screen := screens[0]
	for _, s := range screens {
		if s.IsCurrent {
			screen = s
			break
		}
		if s.IsPrimary {
			screen = s
		}
	}

	w, _ := runtime.WindowGetSize(ctx)
	const margin = 12
	x := screen.Size.Width - w - margin
	y := margin
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	runtime.WindowSetPosition(ctx, x, y)
}
