package data

import (
	"database/sql"
	"time"
)

type DbFile struct {
	ID        string
	Name      string
	Size      int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func GetExistingFileById(db *sql.DB, id string) (*DbFile, error) {
	row := db.QueryRow(`
		SELECT id, name, size, created_at, updated_at
		FROM files
		WHERE id = ? AND deleted_at IS NULL
	`, id)

	var f DbFile
	if err := row.Scan(&f.ID, &f.Name, &f.Size, &f.CreatedAt, &f.UpdatedAt); err != nil {
		return nil, err
	}
	return &f, nil
}

func CreateFile(db *sql.DB, id string, name string, size int64) (*DbFile, error) {
	row := db.QueryRow(`
		INSERT INTO files (id, name, size)
		VALUES (?, ?, ?)
		RETURNING id, name, size, created_at, updated_at;
	`, id, name, size)

	var f DbFile
	if err := row.Scan(&f.ID, &f.Name, &f.Size, &f.CreatedAt, &f.UpdatedAt); err != nil {
		return nil, err
	}
	return &f, nil
}

func GetAllExistingFiles(db *sql.DB) ([]DbFile, error) {
	rows, err := db.Query(`
		SELECT id, name, size, created_at, updated_at
		FROM files
		WHERE deleted_at IS NULL
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := []DbFile{}
	for rows.Next() {
		var f DbFile
		if err := rows.Scan(&f.ID, &f.Name, &f.Size, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, f)
	}

	return result, nil
}

func DeleteExistingFile(db *sql.DB, id string) (bool, error) {
	res, err := db.Exec("UPDATE files SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil {
		return false, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}
