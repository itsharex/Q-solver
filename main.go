package main

import (
	"embed"

	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS


func main() {

	os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGS", "--disable-gpu")
	app := NewApp()
	err := wails.Run(&options.App{
		Title:  "",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		AlwaysOnTop:      true,
		OnStartup:        app.Startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			BackdropType:         windows.None,
			WebviewBrowserPath:   "",
			Theme:                windows.SystemDefault,
		},
		OnShutdown: app.OnShutdown,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
