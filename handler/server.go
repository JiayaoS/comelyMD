package handler

import (
	"io/fs"
	"log"
	"net/http"
	"time"
)

var staticFS fs.FS

func SetStatic(staticAssets fs.FS) {
	staticFS = staticAssets
}

// Run registers routes and starts the HTTP server.
func Run(addr string) {
	mux := http.NewServeMux()

	// Register application routes and static assets.
	mux.HandleFunc("/", IndexHandler)
	mux.HandleFunc("/api/pages", CreatePageHandler)
	mux.HandleFunc("/p/", ViewPageHandler)

	if staticFS != nil {
		fs := http.FileServer(http.FS(staticFS))
		mux.Handle("/static/", http.StripPrefix("/static/", fs))
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Markdown service started: addr=%s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server stopped: %v", err)
	}
}
