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
	"upobir/goriber-drive-v2/internal/web"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite" // CGO-free SQLite
)

const (
	addr         = ":8080"
	publicDir    = "./public"
	storageDir   = "./storage"
	maxUploadMB  = 50
	readTimeout  = 10 * time.Second
	writeTimeout = 30 * time.Second
	idleTimeout  = 60 * time.Second
	dbUrl        = "file:data.db?_pragma=busy_timeout(5000)"
)

func main() {
	if err := os.MkdirAll(storageDir, 0755); err != nil {
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
	r.Use(web.CorsSimple)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api/v1/files", func(r chi.Router) {
		r.Post("/", web.UploadHandler(db, storageDir))
	})

	r.Get("/download/{id}", web.DownloadHandler(db, storageDir))

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
