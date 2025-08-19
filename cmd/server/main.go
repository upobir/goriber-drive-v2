package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"upobir/goriber-drive-v2/internal/service"
	"upobir/goriber-drive-v2/internal/web"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

const (
	// defaultAddr  = ":3333"
	publicDir    = "./public"
	storageDir   = "./storage"
	readTimeout  = 10 * time.Second
	writeTimeout = 30 * time.Second
	idleTimeout  = 60 * time.Second
	dbUrl        = "file:data.db?_pragma=busy_timeout(5000)"
)

func main() {
	port := flag.String("port", "3333", "Port to listen on (overrides default)")
	flag.Parse()

	addr := fmt.Sprintf(":%s", *port)

	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	hub := web.NewHub()

	dependencies := service.Dependencies{
		StorageDir:  storageDir,
		Db:          db,
		Broadcaster: &web.HubBroadcaster{Hub: hub},
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Use(web.CorsSimple)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api/v1/files", func(r chi.Router) {
		r.Get("/", web.GetAllFiles(dependencies))
		r.Post("/", web.UploadHandler(dependencies))
		r.Delete("/{id}", web.DeleteFile(dependencies))
	})

	r.Get("/download/{id}", web.DownloadHandler(dependencies))

	r.Get("/ws", web.ServeWS(hub))

	r.Handle("/*", http.StripPrefix("/", http.FileServer(http.Dir(publicDir))))

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go hub.Run(ctx)

	go func() {
		log.Printf("HTTP server listening at %s", addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)

	log.Println("bye")
}
