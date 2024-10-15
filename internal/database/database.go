package database

import (
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const (
	createTableSql string = `CREATE TABLE IF NOT EXISTS jobs (
		id INTEGER NOT NULL PRIMARY KEY,
		name VARCHAR(50),
		hash VARCHAR(32),
		created DATETIME,
		last_executed DATETIME,
		last_change DATETIME
	);`
)

type Database struct {
	Filename string
	db       *sql.DB
	log      *slog.Logger
}

func NewDb(filename string) *Database {
	return &Database{
		Filename: filename,
		log:      slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (db *Database) Open() error {
	db.log.Info("opening database", "filename", db.Filename)

	var err error
	db.db, err = sql.Open("sqlite3", db.Filename)

	return err
}

func (db *Database) Close() error {
	db.log.Info("closing database", "filename", db.Filename)

	return db.db.Close()
}

func (db *Database) CreateTables() error {
	db.log.Info("creating tables (if not existing)",
		"filename", db.Filename)

	_, err := db.db.Exec(createTableSql)
	return err
}
