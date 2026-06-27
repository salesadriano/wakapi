package repositories

import (
	"testing"
	"time"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeartbeatRepository_InsertBatchAndCount(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "alice")
	repo := NewHeartbeatRepository(db)

	now := time.Now()
	require.NoError(t, repo.InsertBatch([]*models.Heartbeat{
		{UserID: user.ID, Entity: "main.go", Project: "wakapi", Language: "Go", Time: models.CustomTime(now.Add(-2 * time.Minute)), CreatedAt: models.CustomTime(now), Hash: "h1"},
		{UserID: user.ID, Entity: "main.go", Project: "wakapi", Language: "Go", Time: models.CustomTime(now.Add(-1 * time.Minute)), CreatedAt: models.CustomTime(now), Hash: "h2"},
	}))

	count, err := repo.CountByUser(user)
	require.NoError(t, err)
	assert.EqualValues(t, 2, count)

	all, err := repo.GetAll()
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestHeartbeatRepository_InsertBatchDeduplicatesByHash(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "alice")
	repo := NewHeartbeatRepository(db)

	now := time.Now()
	require.NoError(t, repo.InsertBatch([]*models.Heartbeat{
		{UserID: user.ID, Entity: "a.go", Time: models.CustomTime(now), CreatedAt: models.CustomTime(now), Hash: "dup"},
	}))
	// re-inserting the same hash must be a no-op (on-conflict do-nothing)
	require.NoError(t, repo.InsertBatch([]*models.Heartbeat{
		{UserID: user.ID, Entity: "a.go", Time: models.CustomTime(now), CreatedAt: models.CustomTime(now), Hash: "dup"},
	}))

	count, err := repo.CountByUser(user)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
}

func TestHeartbeatRepository_CountByUserIsolatesUsers(t *testing.T) {
	db := newTestDB(t)
	alice := seedUser(t, db, "alice")
	bob := seedUser(t, db, "bob")
	repo := NewHeartbeatRepository(db)

	now := time.Now()
	require.NoError(t, repo.InsertBatch([]*models.Heartbeat{
		{UserID: alice.ID, Entity: "a.go", Time: models.CustomTime(now), CreatedAt: models.CustomTime(now), Hash: "a1"},
		{UserID: alice.ID, Entity: "b.go", Time: models.CustomTime(now), CreatedAt: models.CustomTime(now), Hash: "a2"},
		{UserID: bob.ID, Entity: "c.go", Time: models.CustomTime(now), CreatedAt: models.CustomTime(now), Hash: "b1"},
	}))

	aliceCount, err := repo.CountByUser(alice)
	require.NoError(t, err)
	assert.EqualValues(t, 2, aliceCount)

	bobCount, err := repo.CountByUser(bob)
	require.NoError(t, err)
	assert.EqualValues(t, 1, bobCount)
}
