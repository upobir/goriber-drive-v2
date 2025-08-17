package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

const migrationsDir = "./migrations"

func main() {
	if len(os.Args) < 3 || (os.Args[1] != "up" && os.Args[1] != "down") {
		log.Fatalf("usage: %s (up|down) <num>", os.Args[0])
	}
	up := os.Args[1] == "up"
	target, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("invalid migration number:", err)
	}

	db, err := sql.Open("sqlite", "file:data.db?_pragma=busy_timeout(5000)")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := ensureMigrationsTable(db); err != nil {
		log.Fatal(err)
	}

	current, err := currentVersion(db)
	if err != nil {
		log.Fatal(err)
	}

	versions := getMigrationVersionsToApply(up, current, target)

	if len(versions) == 0 {
		log.Printf("nothing to do, current=%d target=%d", current, target)
		return
	}

	migs, err := loadMigrations(up)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range versions {
		sqlFile, ok := migs[v]
		if !ok {
			log.Fatalf("migration v%d not found", v)
		}
		if up {
			log.Printf("applying migration v%d...", v)
		} else {
			log.Printf("rollig back migration v%d...", v)
		}
		if err := applyMigration(db, v, sqlFile, up); err != nil {
			log.Fatalf("migration v%d failed: %v", v, err)
		}
		if up {
			log.Printf("migration v%d applied", v)
		} else {
			log.Printf("migration v%d rolled back", v)
		}
	}
}

func getMigrationVersionsToApply(up bool, current int, target int) []int {
	result := []int{}
	if up {
		for v := current + 1; v <= target; v++ {
			result = append(result, v)
		}
	} else {
		for v := current; v > target; v-- {
			result = append(result, v)
		}
	}
	return result
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP))`)
	return err
}

func currentVersion(db *sql.DB) (int, error) {
	var v sql.NullInt64
	err := db.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&v)
	if err != nil {
		return 0, err
	}
	if !v.Valid {
		return 0, nil
	}
	return int(v.Int64), nil
}

func loadMigrations(up bool) (map[int]string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	var sqlFile string
	if up {
		sqlFile = "up.sql"
	} else {
		sqlFile = "down.sql"
	}

	migs := make(map[int]string)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "v") {
			continue
		}
		numStr := strings.TrimPrefix(name, "v")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		path := filepath.Join(filepath.Join(migrationsDir, name), sqlFile)
		migs[num] = path
	}
	return migs, nil
}

func applyMigration(db *sql.DB, version int, filePath string, up bool) error {
	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(string(sqlBytes)); err != nil {
		_ = tx.Rollback()
		return err
	}

	var migrationSql string
	if up {
		migrationSql = `INSERT INTO schema_migrations(version) VALUES(?)`
	} else {
		migrationSql = `DELETE FROM schema_migrations where version = ?`
	}

	if _, err := tx.Exec(migrationSql, version); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
