package repositories

import (
	"testing"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiKeyRepository_InsertAndGet(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "alice")
	repo := NewApiKeyRepository(db)

	inserted, err := repo.Insert(&models.ApiKey{
		ApiKey: "secret-key-1",
		UserID: user.ID,
		Label:  "laptop",
	})
	require.NoError(t, err)
	assert.Equal(t, "secret-key-1", inserted.ApiKey)

	got, err := repo.GetByApiKey("secret-key-1", false)
	require.NoError(t, err)
	require.NotNil(t, got.User)
	assert.Equal(t, "alice", got.User.ID)
	assert.Equal(t, "laptop", got.Label)
}

func TestApiKeyRepository_InsertInvalid(t *testing.T) {
	db := newTestDB(t)
	repo := NewApiKeyRepository(db)

	_, err := repo.Insert(&models.ApiKey{ApiKey: "", Label: ""})
	assert.Error(t, err)
}

func TestApiKeyRepository_RequireFullAccessExcludesReadOnly(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "alice")
	repo := NewApiKeyRepository(db)

	_, err := repo.Insert(&models.ApiKey{
		ApiKey:   "readonly-key",
		UserID:   user.ID,
		Label:    "ci",
		ReadOnly: true,
	})
	require.NoError(t, err)

	// a read-only key must not be returned when a full-access key is required
	_, err = repo.GetByApiKey("readonly-key", true)
	assert.Error(t, err)

	// ... but it is returned without that requirement
	got, err := repo.GetByApiKey("readonly-key", false)
	require.NoError(t, err)
	assert.True(t, got.ReadOnly)
}

func TestApiKeyRepository_GetByUserAndDelete(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "alice")
	repo := NewApiKeyRepository(db)

	_, err := repo.Insert(&models.ApiKey{ApiKey: "k1", UserID: user.ID, Label: "a"})
	require.NoError(t, err)
	_, err = repo.Insert(&models.ApiKey{ApiKey: "k2", UserID: user.ID, Label: "b"})
	require.NoError(t, err)

	keys, err := repo.GetByUser(user.ID)
	require.NoError(t, err)
	assert.Len(t, keys, 2)

	require.NoError(t, repo.Delete("k1"))

	keys, err = repo.GetByUser(user.ID)
	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, "k2", keys[0].ApiKey)
}
