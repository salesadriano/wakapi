package routes

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
)

type AdminHandler struct {
	config           *conf.Config
	userService      services.IUserService
	heartbeatService services.IHeartbeatService
}

func NewAdminHandler(userService services.IUserService, heartbeatService services.IHeartbeatService) *AdminHandler {
	return &AdminHandler{
		config:           conf.Get(),
		userService:      userService,
		heartbeatService: heartbeatService,
	}
}

func (h *AdminHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()

	authMiddleware := middlewares.NewAuthenticateMiddleware(h.userService).
		WithRedirectTarget(defaultErrorRedirectTarget()).
		WithRedirectErrorMessage("unauthorized")

	r.Use(authMiddleware.Handler)
	r.Use(h.requireAdmin)

	r.Get("/", h.GetIndex)
	r.Post("/users/{id}/toggle-admin", h.PostToggleAdmin)
	r.Post("/users/{id}/reset-apikey", h.PostResetApiKey)
	r.Post("/users/{id}/delete", h.PostDeleteUser)

	router.Mount("/admin", r)
}

// requireAdmin blocks any request whose principal is not an administrator. It must
// run after the authentication middleware, which populates the principal.
func (h *AdminHandler) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := middlewares.GetPrincipal(r)
		if user == nil || !user.IsAdmin {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("access denied: administrator privileges required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *AdminHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}
	if err := templates[conf.AdminTemplate].Execute(w, h.buildViewModel(r, w)); err != nil {
		conf.Log().Request(r).Error("failed to render admin dashboard", "error", err)
	}
}

func (h *AdminHandler) buildViewModel(r *http.Request, w http.ResponseWriter) *view.AdminViewModel {
	principal := middlewares.GetPrincipal(r)

	users, err := h.userService.GetAll()
	if err != nil {
		conf.Log().Request(r).Error("failed to load users for admin dashboard", "error", err)
		return routeutils.WithSessionMessages(&view.AdminViewModel{
			SharedLoggedInViewModel: view.SharedLoggedInViewModel{
				SharedViewModel: view.NewSharedViewModel(h.config, &view.Messages{Error: criticalError}),
				User:            principal,
			},
		}, r, w)
	}

	// heartbeat counts per user (single batch query, avoids N+1)
	countByUser := map[string]int64{}
	if counts, err := h.heartbeatService.CountByUsers(users); err == nil {
		for _, c := range counts {
			countByUser[c.User] = c.Count
		}
	} else {
		conf.Log().Request(r).Warn("failed to count heartbeats per user", "error", err)
	}

	// last activity per user (single batch query)
	lastByUser := map[string]time.Time{}
	if last, err := h.heartbeatService.GetLastAll(); err == nil {
		for _, t := range last {
			lastByUser[t.User] = t.Time.T()
		}
	} else {
		conf.Log().Request(r).Warn("failed to fetch last activity per user", "error", err)
	}

	entries := make([]*view.AdminUserEntry, 0, len(users))
	for _, u := range users {
		entries = append(entries, &view.AdminUserEntry{
			User:           u,
			HeartbeatCount: countByUser[u.ID],
			LastSeen:       lastByUser[u.ID],
		})
	}

	// rank developers by activity (heartbeat count) desc, then by id for stable order
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].HeartbeatCount != entries[j].HeartbeatCount {
			return entries[i].HeartbeatCount > entries[j].HeartbeatCount
		}
		return entries[i].User.ID < entries[j].User.ID
	})

	totalHeartbeats, _ := h.heartbeatService.Count(true)
	online, _ := h.userService.CountCurrentlyOnline()

	vm := &view.AdminViewModel{
		SharedLoggedInViewModel: view.SharedLoggedInViewModel{
			SharedViewModel: view.NewSharedViewModel(h.config, nil),
			User:            principal,
		},
		Users:           entries,
		TotalUsers:      len(users),
		TotalHeartbeats: totalHeartbeats,
		OnlineUsers:     online,
	}
	return routeutils.WithSessionMessages(vm, r, w)
}

func (h *AdminHandler) PostToggleAdmin(w http.ResponseWriter, r *http.Request) {
	principal := middlewares.GetPrincipal(r)
	target, ok := h.resolveTarget(w, r)
	if !ok {
		return
	}
	if principal != nil && target.ID == principal.ID {
		routeutils.SetError(r, w, "you cannot change your own admin status")
		h.redirectToDashboard(w, r)
		return
	}

	target.IsAdmin = !target.IsAdmin
	if _, err := h.userService.Update(target); err != nil {
		conf.Log().Request(r).Error("admin: failed to update user", "userID", target.ID, "error", err)
		routeutils.SetError(r, w, conf.ErrInternalServerError)
		h.redirectToDashboard(w, r)
		return
	}

	routeutils.SetSuccess(r, w, fmt.Sprintf("updated admin status for '%s'", target.ID))
	h.redirectToDashboard(w, r)
}

func (h *AdminHandler) PostResetApiKey(w http.ResponseWriter, r *http.Request) {
	target, ok := h.resolveTarget(w, r)
	if !ok {
		return
	}
	if _, err := h.userService.ResetApiKey(target); err != nil {
		conf.Log().Request(r).Error("admin: failed to reset api key", "userID", target.ID, "error", err)
		routeutils.SetError(r, w, conf.ErrInternalServerError)
		h.redirectToDashboard(w, r)
		return
	}

	routeutils.SetSuccess(r, w, fmt.Sprintf("reset API key for '%s'", target.ID))
	h.redirectToDashboard(w, r)
}

func (h *AdminHandler) PostDeleteUser(w http.ResponseWriter, r *http.Request) {
	principal := middlewares.GetPrincipal(r)
	target, ok := h.resolveTarget(w, r)
	if !ok {
		return
	}
	if principal != nil && target.ID == principal.ID {
		routeutils.SetError(r, w, "you cannot delete your own account from the admin dashboard")
		h.redirectToDashboard(w, r)
		return
	}
	if err := h.userService.Delete(target); err != nil {
		conf.Log().Request(r).Error("admin: failed to delete user", "userID", target.ID, "error", err)
		routeutils.SetError(r, w, conf.ErrInternalServerError)
		h.redirectToDashboard(w, r)
		return
	}

	routeutils.SetSuccess(r, w, fmt.Sprintf("deleted user '%s' and all associated data", target.ID))
	h.redirectToDashboard(w, r)
}

// resolveTarget loads the user referenced by the {id} route param, redirecting with
// an error flash if it does not exist.
func (h *AdminHandler) resolveTarget(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	targetID := chi.URLParam(r, "id")
	target, err := h.userService.GetUserById(targetID)
	if err != nil || target == nil {
		routeutils.SetError(r, w, "user not found")
		h.redirectToDashboard(w, r)
		return nil, false
	}
	return target, true
}

func (h *AdminHandler) redirectToDashboard(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, h.config.Server.BasePath+"/admin", http.StatusFound)
}
