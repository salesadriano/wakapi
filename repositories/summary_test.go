package repositories

import (
	"testing"
	"time"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummaryRepository_InsertAndGetAll(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "alice")
	repo := NewSummaryRepository(db)

	from := time.Now().Add(-2 * time.Hour)
	to := time.Now().Add(-1 * time.Hour)
	require.NoError(t, repo.Insert(&models.Summary{
		UserID:   user.ID,
		FromTime: models.CustomTime(from),
		ToTime:   models.CustomTime(to),
		Languages: models.SummaryItems{
			{Type: models.SummaryLanguage, Key: "Go", Total: 30 * time.Minute},
		},
	}))

	all, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, user.ID, all[0].UserID)
	assert.NotZero(t, all[0].ID)
}

func TestSummaryRepository_GetLastBySingleUser(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "alice")
	repo := NewSummaryRepository(db)

	// no summaries yet -> zero time, no error
	got, err := repo.GetLastBySingleUser(user.ID)
	require.NoError(t, err)
	assert.True(t, got.IsZero())

	base := time.Now()
	require.NoError(t, repo.Insert(&models.Summary{
		UserID:   user.ID,
		FromTime: models.CustomTime(base.Add(-2 * time.Hour)),
		ToTime:   models.CustomTime(base.Add(-90 * time.Minute)),
	}))
	require.NoError(t, repo.Insert(&models.Summary{
		UserID:   user.ID,
		FromTime: models.CustomTime(base.Add(-1 * time.Hour)),
		ToTime:   models.CustomTime(base.Add(-30 * time.Minute)),
	}))

	got, err = repo.GetLastBySingleUser(user.ID)
	require.NoError(t, err)
	assert.False(t, got.IsZero())
}

func TestSummaryRepository_DeleteByUser(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "alice")
	repo := NewSummaryRepository(db)

	require.NoError(t, repo.Insert(&models.Summary{
		UserID:   user.ID,
		FromTime: models.CustomTime(time.Now().Add(-time.Hour)),
		ToTime:   models.CustomTime(time.Now()),
	}))

	require.NoError(t, repo.DeleteByUser(user.ID))

	all, err := repo.GetAll()
	require.NoError(t, err)
	assert.Empty(t, all)
}
