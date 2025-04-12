package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("error pinging database: %v", err)
	}

	err = createTables()
	if err != nil {
		return fmt.Errorf("error creating tables: %v", err)
	}

	return nil
}

func GetDB() *sql.DB {
	return db
}

func createTables() error {
	// Create tables with a single transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	// Create users table
	_, err = tx.Exec(`
    CREATE TABLE IF NOT EXISTS users (
      id SERIAL PRIMARY KEY,
      username VARCHAR(255) UNIQUE NOT NULL,
      password_hash VARCHAR(255) NOT NULL,
      email TEXT,
      did TEXT UNIQUE,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
      last_login_at TIMESTAMP WITH TIME ZONE,
      last_updated_at TIMESTAMP WITH TIME ZONE,
      is_two_factor_enabled BOOLEAN DEFAULT false
    )
  `)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	// Create sessions table
	_, err = tx.Exec(`
    CREATE TABLE IF NOT EXISTS sessions (
      id SERIAL PRIMARY KEY,
      user_id INTEGER REFERENCES users(id),
      token TEXT UNIQUE NOT NULL,
      expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    )
  `)
	if err != nil {
		return fmt.Errorf("error creating sessions table: %v", err)
	}

	// Update any existing records that might have null timestamps
	_, err = tx.Exec(`
    UPDATE users 
    SET 
      last_login_at = COALESCE(last_login_at, created_at),
      last_updated_at = COALESCE(last_updated_at, created_at),
      is_two_factor_enabled = COALESCE(is_two_factor_enabled, false)
    WHERE 
      last_login_at IS NULL 
      OR last_updated_at IS NULL 
      OR is_two_factor_enabled IS NULL
  `)
	if err != nil {
		return fmt.Errorf("error updating existing records: %v", err)
	}

	// Create indexes for better query performance
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users (username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_did ON users (did)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions (token)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id)`,
	}

	for _, idx := range indexes {
		_, err = tx.Exec(idx)
		if err != nil {
			return fmt.Errorf("error creating index: %v", err)
		}
	}

	// Commit all changes
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing changes: %v", err)
	}

	return nil
}
