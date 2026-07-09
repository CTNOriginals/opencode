package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func OpenWriter(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_txlock=immediate&synchronous=NORMAL&_busy_timeout=5000&_foreign_keys=ON", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func OpenReader(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&synchronous=NORMAL&_busy_timeout=5000&_foreign_keys=ON", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	return db, nil
}
