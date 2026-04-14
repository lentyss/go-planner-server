package db

import (
	"database/sql"
	_ "embed"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var (
	//go:embed 001_create_scheduler.sql
	schemaSQL string
	DB        *sql.DB
)

func Init(dbFile string) (*sql.DB, error) {
	dir := filepath.Dir(dbFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	_, err := os.Stat(dbFile)
	install := err != nil

	DB, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}

	if err = DB.Ping(); err != nil {
		DB.Close()
		return nil, err
	}

	if install {
		log.Printf("Database file %s not found, running migrations...", dbFile)

		if _, err = DB.Exec(schemaSQL); err != nil {
			DB.Close()
			return nil, err
		}
	}

	log.Printf("Database %s is ready", dbFile)
	return DB, nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
