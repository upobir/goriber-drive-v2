package service

import (
	"errors"
	"fmt"
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

func fromDB(f data.DbFile) File {
	return File{
		ID:        f.ID,
		Name:      f.Name,
		Size:      f.Size,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
		DeletedAt: f.DeletedAt,
	}
}

func SaveFile(dependencies Dependencies, file io.Reader, name string) (*File, error) {
	id := uuid.NewString()
	storedPath := filepath.Join(dependencies.StorageDir, id)

	dst, err := os.Create(storedPath)
	if err != nil {
		return nil, ErrDiskWriteFailure
	}

	size, err := io.Copy(dst, file)
	if err != nil {
		return nil, ErrDiskWriteFailure
	}

	dst.Close()

	dbFile, err := data.CreateFile(dependencies.Db, id, name, size)
	if err != nil {
		return nil, ErrUnknownDbError
	}

	fileModel := fromDB(*dbFile)
	dependencies.Broadcaster.Broadcast(Event{
		Type:    EventFileCreated,
		Payload: fileModel,
	})
	return &fileModel, nil
}

func GetFileWithOsFileById(dependencies Dependencies, id string) (*FileWithOsFile, error) {
	dbFile, err := data.GetExistingFileById(dependencies.Db, id)
	if err != nil {
		return nil, ErrUnknownDbError
	}
	if dbFile == nil {
		return nil, ErrNotFoundInDB
	}

	file := fromDB(*dbFile)

	path := filepath.Join(dependencies.StorageDir, id)
	osFile, err := os.Open(path)
	if err != nil {
		return nil, ErrNotFoundInDisk
	}

	return &FileWithOsFile{
		File:   file,
		OsFile: osFile,
	}, nil
}

func GetAllFiles(dependencies Dependencies) ([]File, error) {
	dbFiles, err := data.GetAllExistingFiles(dependencies.Db)
	if err != nil {
		fmt.Println("error: ", err)
		return nil, ErrUnknownDbError
	}

	result := []File{}
	for _, f := range dbFiles {
		result = append(result, fromDB(f))
	}

	return result, nil
}

func DeleteFile(dependencies Dependencies, id string) error {
	deleted, err := data.DeleteExistingFile(dependencies.Db, id)
	if err != nil {
		fmt.Println("error: ", err)
		return ErrUnknownDbError
	}

	if !deleted {
		return ErrNotFoundInDB
	}

	path := filepath.Join(dependencies.StorageDir, id)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return ErrDiskWriteFailure
	}

	dependencies.Broadcaster.Broadcast(Event{
		Type:    EventFileDeleted,
		Payload: id,
	})

	return nil
}
