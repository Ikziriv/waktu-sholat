//go:build !desktop

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pkg/browser"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "check" {
		if err := runSelfCheck(); err != nil {
			fmt.Fprintln(os.Stderr, "FAIL:", err)
			os.Exit(1)
		}
		fmt.Println("self-check ok")
		return
	}
	defaultAddr := "0.0.0.0:9090"
	if p := os.Getenv("PORT"); p != "" {
		defaultAddr = "0.0.0.0:" + p
	}
	addr := flag.String("addr", defaultAddr, "listen address")
	flag.Parse()

	mux := buildMux()
	log.Printf("waktu-sholat listening on %s", *addr)
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = browser.OpenURL("http://" + *addr)
	}()
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}
