package repositories

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newTestDB spins up an isolated, file-backed SQLite database (pure-Go driver,
// works with CGO_ENABLED=0) and auto-migrates the subset of models exercised by
// the repository tests. Each test gets its own database via t.TempDir().
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// CustomTime.Scan reads config.Get() (postgres timezone hack); without an
	// initialized config it would nil-panic on the first time-valued column.
	if config.Get() == nil {
		config.Set(config.Empty())
	}

	dbFile := filepath.Join(t.TempDir(), "wakapi_test.db")
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	require.NoError(t, err, "failed to open test database")

	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.WebAuthnCredential{},
		&models.ApiKey{},
		&models.Heartbeat{},
		&models.Summary{},
		&models.SummaryItem{},
	), "failed to migrate test schema")

	return db
}

// seedUser inserts a minimal, valid user and returns it.
func seedUser(t *testing.T, db *gorm.DB, id string) *models.User {
	t.Helper()

	u := &models.User{
		ID:       id,
		ApiKey:   id + "-key",
		Email:    id + "@example.test",
		Password: "hashed-password",
	}
	require.NoError(t, db.Create(u).Error, "failed to seed user %q", id)
	return u
}
