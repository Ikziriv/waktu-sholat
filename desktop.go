//go:build desktop

package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/dist

var frontendAssets embed.FS

// App is the Wails application struct. Methods on it are callable from JS.
type App struct {
	ctx context.Context
}

// OnStartup starts the Go HTTP server in a goroutine when the Wails app launches.
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	go func() {
		mux := buildMux()
		addr := "127.0.0.1:9090"
		log.Printf("waktu-sholat desktop server on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal(err)
		}
	}()
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "check" {
		if err := runSelfCheck(); err != nil {
			fmt.Fprintln(os.Stderr, "FAIL:", err)
			os.Exit(1)
		}
		fmt.Println("self-check ok")
		return
	}

	err := wails.Run(&options.App{
		Title:  "Waktu Sholat",
		Width:  800,
		Height: 700,
		AssetServer: &assetserver.Options{
			Assets: frontendAssets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 14, B: 26, A: 1},
		OnStartup:        (&App{}).OnStartup,
		Bind:             []interface{}{&App{}},
	})
	if err != nil {
		log.Fatal(err)
	}
}
