package db

import (
	"testing"
)

func TestNewConnection_ReturnsDB(t *testing.T) {
	db, err := NewConnection(":memory:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if db == nil {
		t.Fatal("expected db to be non-nil")
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("expected no error getting sql.DB, got %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("expected ping to succeed, got %v", err)
	}
}
