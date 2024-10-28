package database

import (
	"database/sql"
	"fmt"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"vmuser/config"
)

func GetConnection(cfg *config.Turso) (*sql.DB, error) {
	db, err := sql.Open("libsql", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("error opening connection: %w", err)
	}
	return db, nil
}
