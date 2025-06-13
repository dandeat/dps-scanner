package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	// Create directory if it doesn't exist
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		err := os.Mkdir("data", 0755)
		if err != nil {
			log.Fatalf("failed to create data dir: %v", err)
		}
	}

	database, err := sql.Open("sqlite3", "./data/scanner.db")
	if err != nil {
		log.Fatalf("failed to open SQLite database: %v", err)
	}

	// Set connection pool settings (optional for SQLite)
	database.SetMaxOpenConns(1)

	// Test connection
	if err := database.Ping(); err != nil {
		log.Fatalf("failed to ping SQLite database: %v", err)
	}

	fmt.Println("Connected to SQLite successfully.")
	DB = database

	createTables()
}

func createTables() {
	createScanHistory := `
	CREATE TABLE IF NOT EXISTS scan_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT,
		muat_id TEXT,
		barcode TEXT,
		ip_address TEXT,
		location TEXT,
		scanned_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := DB.Exec(createScanHistory)
	if err != nil {
		log.Fatalf("failed to create scan_history table: %v", err)
	}
}
