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
	// search & pagination
	Search       string
	Page         int
	PageSize     int
	TotalMatched int // number of users matching the current search (pre-pagination)
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

func (m *AdminViewModel) TotalPages() int {
	if m.PageSize <= 0 {
		return 1
	}
	n := (m.TotalMatched + m.PageSize - 1) / m.PageSize
	if n < 1 {
		return 1
	}
	return n
}

func (m *AdminViewModel) HasPrev() bool { return m.Page > 1 }

func (m *AdminViewModel) HasNext() bool { return m.Page < m.TotalPages() }

func (m *AdminViewModel) PrevPage() int {
	if m.Page > 1 {
		return m.Page - 1
	}
	return 1
}

func (m *AdminViewModel) NextPage() int {
	if m.Page < m.TotalPages() {
		return m.Page + 1
	}
	return m.TotalPages()
}
