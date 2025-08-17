package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite" // CGO-free SQLite
)

const (
	addr         = ":8080"
	publicDir    = "./public"
	uploadDir    = "./storage"
	maxUploadMB  = 50
	readTimeout  = 10 * time.Second
	writeTimeout = 30 * time.Second
	idleTimeout  = 60 * time.Second
	dbUrl        = "file:data.db?_pragma=busy_timeout(5000)"
)

func corsSimple(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// if err := initSchema(db); err != nil {
	// 	log.Fatal(err)
	// }

	// hub := newHub()
	// go hub.run()

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Use(corsSimple)

	// Health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Handle("/*", http.StripPrefix("/", http.FileServer(http.Dir(publicDir))))

	// r.Route("/api", func(api chi.Router) {
	// 	api.Get("/notes", listNotes(db))
	// 	api.Post("/notes", createNote(db, hub))

	// 	api.Post("/upload", uploadFile(db, hub))
	// 	api.Get("/files", listFiles(db))
	// 	api.Get("/files/{stored}", downloadFile(db)) // download by stored filename
	// })

	// WebSocket for realtime events
	// r.Get("/ws", wsHandler(hub))

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// graceful shutdown
	go func() {
		log.Printf("HTTP server listening at %s", addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("bye")
}
