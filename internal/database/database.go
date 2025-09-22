package database

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// NewDB creates a new database connection pool.
func NewDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")
	return db, nil
}

// Migrate applies all pending database migrations.
func Migrate(dsn string) error {
	m, err := migrate.New(
		"file://internal/database/migrations",
		fmt.Sprintf("mysql://%s", dsn),
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

const (
	// baseDSN should not include a database name, e.g., "user:password@tcp(host:port)/"
	baseDSN = "root:password@tcp(127.0.0.1:3306)/?parseTime=true&allowNativePasswords=true"
	// migrationsPath is the path to the migration files.
	migrationsPath = "file://internal/database/migrations"
	// testMigrationsPath is for tests running in other packages.
	testMigrationsPath = "file://../database/migrations"
)

// SetupTestDB creates a new, random database for testing, runs migrations, and returns the connection.
func SetupTestDB(t *testing.T) (*sql.DB, string) {
	t.Helper()

	// 1. Create a random database name
	dbName := fmt.Sprintf("utopia_test_%d_%d", time.Now().UnixNano(), rand.Intn(1000))

	// 2. Connect to MySQL server (without a specific database)
	db, err := sql.Open("mysql", baseDSN)
	if err != nil {
		t.Fatalf("failed to connect to mysql for setup: %v", err)
	}
	defer db.Close()

	// 3. Create the new database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		t.Fatalf("failed to create test database %s: %v", dbName, err)
	}

	// 4. Connect to the newly created test database
	testDSNWithDB := fmt.Sprintf("root:password@tcp(127.0.0.1:3306)/%s?parseTime=true", dbName)
	testDB, err := NewDB(testDSNWithDB)
	if err != nil {
		t.Fatalf("failed to connect to test database %s: %v", dbName, err)
	}

	// 5. Run migrations
	migrateDSN := fmt.Sprintf("mysql://%s", testDSNWithDB)
	m, err := migrate.New(
		testMigrationsPath,
		migrateDSN,
	)
	if err != nil {
		// Cleanup before failing
		testDB.Close()
		TeardownTestDB(nil, dbName) // Pass nil for db as it's already closed
		t.Fatalf("failed to create migrate instance for test db: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		// Cleanup before failing
		testDB.Close()
		TeardownTestDB(nil, dbName)
		t.Fatalf("failed to apply migrations to test db: %v", err)
	}

	log.Printf("Successfully created and migrated test database: %s", dbName)
	return testDB, dbName
}

// TeardownTestDB drops the test database and closes the connection.
func TeardownTestDB(db *sql.DB, dbName string) {
	if db != nil {
		db.Close()
	}

	// Connect to MySQL server again to drop the database
	rootDB, err := sql.Open("mysql", baseDSN)
	if err != nil {
		log.Printf("failed to connect to mysql for teardown: %v", err)
		return
	}
	defer rootDB.Close()

	_, err = rootDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		log.Printf("failed to drop test database %s: %v", dbName, err)
	} else {
		log.Printf("Successfully dropped test database: %s", dbName)
	}
}
