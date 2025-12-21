package main

import (
	"embed"
	"log/slog"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

// During development, assets are served from the dev server
// During production build, wails embeds the frontend/dist folder
var assets embed.FS

func main() {
	// ファイルロガーを初期化（ターミナル + ファイル出力）
	logResult, err := logging.NewFileLogger(logging.FileLoggerConfig{
		FilePrefix: "multiverse",
		Config:     logging.DebugConfig(),
	})
	if err != nil {
		println("Failed to initialize logger:", err.Error())
	} else {
		// デフォルトロガーとして設定
		slog.SetDefault(logResult.Logger)
		slog.Info("Multiverse IDE starting", "log_file", logResult.LogFilePath)
		defer func() {
			if closeErr := logResult.Close(); closeErr != nil {
				slog.Error("Failed to close log file", "error", closeErr)
			}
		}()
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "Multiverse",
		Width:  1920,
		Height: 1080,
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
