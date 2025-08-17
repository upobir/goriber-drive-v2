package service

import (
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
	"upobir/goriber-drive-v2/internal/data"

	"github.com/google/uuid"
)

type File struct {
	ID        string
	Name      string
	Size      int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type FileWithOsFile struct {
	File   File
	OsFile *os.File
}

var (
	ErrNotFoundInDB     = errors.New("not found in database")
	ErrNotFoundInDisk   = errors.New("not found in disk storage")
	ErrUnknownDbError   = errors.New("something wrong in data layer")
	ErrDiskWriteFailure = errors.New("failed to write to disk")
)

func SaveFile(db *sql.DB, storageDir string, file io.Reader, name string) (*File, error) {
	id := uuid.NewString()
	storedPath := filepath.Join(storageDir, id)

	dst, err := os.Create(storedPath)
	if err != nil {
		return nil, ErrDiskWriteFailure
	}

	size, err := io.Copy(dst, file)
	if err != nil {
		return nil, ErrDiskWriteFailure
	}

	dst.Close()

	dbFile, err := data.CreateFile(db, id, name, size)
	if err != nil {
		return nil, ErrUnknownDbError
	}

	return &File{
		ID:        dbFile.ID,
		Name:      dbFile.Name,
		Size:      dbFile.Size,
		CreatedAt: dbFile.CreatedAt,
		UpdatedAt: dbFile.UpdatedAt,
		DeletedAt: dbFile.DeletedAt,
	}, nil
}

func GetFileWithOsFileById(db *sql.DB, storageDir string, id string) (*FileWithOsFile, error) {
	dbFile, err := data.GetExistingFileById(db, id)
	if err != nil {
		return nil, ErrUnknownDbError
	}
	if dbFile == nil {
		return nil, ErrNotFoundInDB
	}

	file := File{
		ID:        dbFile.ID,
		Name:      dbFile.Name,
		Size:      dbFile.Size,
		CreatedAt: dbFile.CreatedAt,
		UpdatedAt: dbFile.UpdatedAt,
		DeletedAt: dbFile.DeletedAt,
	}

	path := filepath.Join(storageDir, id)
	osFile, err := os.Open(path)
	if err != nil {
		return nil, ErrNotFoundInDisk
	}

	return &FileWithOsFile{
		File:   file,
		OsFile: osFile,
	}, nil
}
