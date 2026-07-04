package database

import (
	"testing"
)

func TestConfig_DSN(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	got := cfg.DSN()

	if got != expected {
		t.Errorf("DSN() = %q, want %q", got, expected)
	}
}

func TestConfig_DSN_EmptyPassword(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "inventory",
		Password: "",
		DBName:   "personal_digital_inventory",
		SSLMode:  "disable",
	}

	expected := "host=localhost port=5432 user=inventory password= dbname=personal_digital_inventory sslmode=disable"
	got := cfg.DSN()

	if got != expected {
		t.Errorf("DSN() = %q, want %q", got, expected)
	}
}

func TestConnect_InvalidHost(t *testing.T) {
	cfg := Config{
		Host:     "invalid-host-that-does-not-exist",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	db, err := Connect(cfg)
	if err == nil {
		_ = Close(db)
		t.Fatal("Connect() should return error for invalid host")
	}

	if db != nil {
		t.Error("Connect() should return nil db on error")
	}
}

func TestConnect_InvalidPort(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     1,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	db, err := Connect(cfg)
	if err == nil {
		_ = Close(db)
		t.Fatal("Connect() should return error for invalid port")
	}

	if db != nil {
		t.Error("Connect() should return nil db on error")
	}
}
