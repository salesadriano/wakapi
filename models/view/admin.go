package view

import (
	"time"

	"github.com/muety/wakapi/models"
)

type AdminViewModel struct {
	SharedLoggedInViewModel
	Users           []*AdminUserEntry
	TotalUsers      int
	TotalHeartbeats int64
	OnlineUsers     int
}

// AdminUserEntry is a single developer row in the admin dashboard, enriched with
// server-side aggregated activity stats.
type AdminUserEntry struct {
	User           *models.User
	HeartbeatCount int64
	LastSeen       time.Time
}

func (e *AdminUserEntry) LastSeenLabel() string {
	if e.LastSeen.IsZero() {
		return "never"
	}
	return e.LastSeen.Format("2006-01-02 15:04")
}
