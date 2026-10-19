package database

import (
	"os"
	"strings"
	"testing"
)

func TestConnectMissingUser(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("DB_USER", "")
	if err := Connect(); err == nil || !strings.Contains(err.Error(), "DB_USER") {
		t.Fatalf("missing user must fail before connecting: %v", err)
	}
}

func TestConnectLoadsDotEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	// เก็บค่าเดิมให้ Cleanup คืน แล้วนำออกเพื่อให้ godotenv โหลดจากไฟล์ได้
	for _, key := range []string{"DB_USER", "DB_PORT"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(".env", []byte("DB_USER=from_file\nDB_PORT=invalid\n"), 0600); err != nil {
		t.Fatal(err)
	}
	// ต้องผ่านการอ่าน DB_USER แล้วล้มเหลวที่ Port โดยไม่เชื่อมฐานข้อมูล
	if err := Connect(); err == nil || !strings.Contains(err.Error(), "DB_PORT") {
		t.Fatalf(".env was not loaded correctly: %v", err)
	}
	if os.Getenv("DB_USER") != "from_file" {
		t.Fatal("expected DB_USER loaded from .env")
	}
}

func TestConnectEnvironmentOverridesDotEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("DB_USER", "from_environment")
	t.Setenv("DB_PORT", "invalid")
	if err := os.WriteFile(".env", []byte("DB_USER=\nDB_PORT=3306\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Connect(); err == nil || !strings.Contains(err.Error(), "DB_PORT") {
		t.Fatalf("environment must take precedence: %v", err)
	}
}

func TestConnectInvalidPort(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("DB_USER", "workshop")
	for _, port := range []string{"0", "65536", "invalid"} {
		t.Setenv("DB_PORT", port)
		if err := Connect(); err == nil || !strings.Contains(err.Error(), "DB_PORT") {
			t.Errorf("invalid port %q accepted: %v", port, err)
		}
	}
}
