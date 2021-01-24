package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // Ensure sqlite3 is active
)

// A Store stores some public data.
type Store struct {
	SqliteStore
}

func (s *SqliteStore) Close() error {
	return s.db.Close()
}

func (s *SqliteStore) configValue(key string) (string, error) {
	rows, err := s.db.Query("SELECT content FROM config WHERE key=?;", key)
	if err != nil {
		return "", err
	}
	content := ""
	defer rows.Close()

	for rows.Next() {
		if content != "" {
			return "", fmt.Errorf("More than one result for '%s' in config", key)
		}
		err = rows.Scan(&content)
		if err != nil {
			return "", err
		}
	}
	err = rows.Err()
	if err != nil {
		return "", err
	}
	if content == "" {
		return "", fmt.Errorf("No key '%s' in config table", key)
	}
	return content, nil
}

func (s *SqliteStore) setConfigValue(key string, content string) error {
	res, err := s.db.Exec("INSERT OR REPLACE INTO config (key, content) VALUES (?, ?);", key, content)
	fmt.Printf("Got %s from setting\n", res)
	return err
}

// Workspace Return the name of the workspace
func (s *SqliteStore) Workspace() (string, error) {
	return s.configValue("workspace")
}

// SqliteStore is the primary store backing
type SqliteStore struct {
	db *sql.DB
}

// Open opens a new file :)
func (s *SqliteStore) Open(filename string, workspace string) error {
	_, err := s.init(filename, workspace)
	return err
}

func migrate(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	createStatements := []string{
		`CREATE TABLE IF NOT EXISTS
		config (
                key TEXT NOT NULL PRIMARY KEY,
                content TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS
		docs (
            format TEXT NOT NULL,
            workspace TEXT NOT NULL,
            path TEXT NOT NULL,
            contentHash TEXT NOT NULL,
            content TEXT NOT NULL, -- TODO: allow null
            author TEXT NOT NULL,
            timestamp NUMBER NOT NULL,
            deleteAfter NUMBER,  -- can be null
            signature TEXT NOT NULL,
            PRIMARY KEY(path, author)
		)`,
		"CREATE INDEX IF NOT EXISTS idx1 ON docs(path, author)",
		"CREATE INDEX IF NOT EXISTS idx2 ON docs(path, timestamp)",
		"CREATE INDEX IF NOT EXISTS idx3 ON docs(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx4 ON docs(author)",
	}

	for _, stmt := range createStatements {
		_, err = db.Exec(stmt)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Init gets it all ready to go
// Creates the file if necessary, creates tables, etc.
func (s *SqliteStore) init(filename string, workspace string) (db *sql.DB, err error) {

	db, err = sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	s.db = db
	err = migrate(db)
	if err != nil {
		return nil, err
	}
	err = s.setConfigValue("workspace", workspace)
	if err != nil {
		return nil, err
	}

	fmt.Println("it wekred")
	return db, nil
}
