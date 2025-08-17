package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"upobir/goriber-drive-v2/internal/service"

	"github.com/go-chi/chi/v5"
)

const (
	MaxFileSize = 5 << 30
)

type FileResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func CorsSimple(next http.Handler) http.Handler {
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

func UploadHandler(db *sql.DB, storageDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)

		if err := r.ParseMultipartForm(MaxFileSize); err != nil {
			http.Error(w, "invalid form: "+err.Error(), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file: "+err.Error(), http.StatusBadRequest)
			return
		}

		defer file.Close()
		name := header.Filename

		fileModel, err := service.SaveFile(db, storageDir, file, name)
		if err != nil {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		fileResponse := FileResponse{
			ID:        fileModel.ID,
			Name:      fileModel.Name,
			Size:      fileModel.Size,
			CreatedAt: fileModel.CreatedAt,
			UpdatedAt: fileModel.UpdatedAt,
		}

		jsonOK(w, fileResponse)
	}
}

func DownloadHandler(db *sql.DB, storageDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		file, err := service.GetFileWithOsFileById(db, storageDir, id)
		switch {
		case errors.Is(err, service.ErrNotFoundInDB):
			http.Error(w, "not found", http.StatusNotFound)
			return
		case errors.Is(err, service.ErrNotFoundInDisk):
			http.Error(w, "file missing", http.StatusInternalServerError)
			return
		case err != nil:
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		defer file.OsFile.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.File.Name))

		http.ServeContent(w, r, file.File.Name, time.Now(), file.OsFile)
	}
}
