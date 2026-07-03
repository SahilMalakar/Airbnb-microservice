package config

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// LoadDb opens a connection pool to Postgres and verifies it's reachable.
// The caller owns the returned *sql.DB and is responsible for closing it
// when the application shuts down.
func LoadDb() (*sql.DB, error) {
	dsn := GetEnvString("DATABASE_URL", "")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	fmt.Println("we are connecting to database ...")

	// db.Open() only creates a connection pool, and doesn't actually establish
	// a connection. To ensure the connection works you need to do *something*
	// with a connection.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("successfully connected to database")

	return db, nil
}