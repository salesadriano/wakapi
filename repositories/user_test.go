package repositories

import (
	"testing"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_FindOneByApiKey(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	u := seedUser(t, db, "alice")

	got, err := repo.FindOne(models.User{ApiKey: u.ApiKey})
	require.NoError(t, err)
	assert.Equal(t, "alice", got.ID)
	assert.Equal(t, u.ApiKey, got.ApiKey)
}

func TestUserRepository_FindOneNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.FindOne(models.User{ApiKey: "does-not-exist"})
	assert.Error(t, err)
}

func TestUserRepository_GetByIds(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	seedUser(t, db, "alice")
	seedUser(t, db, "bob")
	seedUser(t, db, "carol")

	users, err := repo.GetByIds([]string{"alice", "carol"})
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserRepository_GetAll(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	seedUser(t, db, "alice")
	seedUser(t, db, "bob")

	users, err := repo.GetAll()
	require.NoError(t, err)
	assert.Len(t, users, 2)
}
