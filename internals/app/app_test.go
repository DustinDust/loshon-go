package app

import (
	"testing"
)

func TestCreateApp(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("SECRET_PATH", "./../../.env.test")
	// due to the way the test is structured, the .env.test file is not loaded correctly
	app := NewApp()
	if app.config.Environment != "test" {
		t.Errorf("Expected test environment, got %s", app.config.Environment)
	}
	if app.db.Migrator().CurrentDatabase() != "loshon_test" {
		t.Errorf("Expected loshon_test database, got %s", app.db.Migrator().CurrentDatabase())
	}
}
