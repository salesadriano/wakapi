package routes

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newAdminTestHandler(userMock *mocks.UserServiceMock) *AdminHandler {
	return &AdminHandler{
		config:           conf.Get(),
		userService:      userMock,
		heartbeatService: new(mocks.HeartbeatServiceMock),
	}
}

func injectPrincipal(principal *models.User) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if principal != nil {
				routeutils.SetPrincipal(r, principal)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// adminActionRouter wires the SharedData + principal injection + admin gate around
// the POST action routes (no GET, to avoid template rendering in unit tests).
func adminActionRouter(principal *models.User, h *AdminHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.NewSharedDataMiddleware())
	r.Use(injectPrincipal(principal))
	r.Use(h.requireAdmin)
	r.Post("/admin/users", h.PostCreateUser)
	r.Post("/admin/users/{id}/edit", h.PostEditUser)
	r.Post("/admin/users/{id}/reset-password", h.PostResetPassword)
	r.Post("/admin/users/{id}/toggle-admin", h.PostToggleAdmin)
	r.Post("/admin/users/{id}/reset-apikey", h.PostResetApiKey)
	r.Post("/admin/users/{id}/delete", h.PostDeleteUser)
	return r
}

func postForm(target string, values url.Values) *http.Request {
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// The headline requirement: the dashboard must be reachable by administrators only.
func TestAdminHandler_RequireAdmin(t *testing.T) {
	conf.Set(conf.Empty())
	h := newAdminTestHandler(new(mocks.UserServiceMock))

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cases := []struct {
		name      string
		principal *models.User
		want      int
	}{
		{"admin is allowed", &models.User{ID: "root", IsAdmin: true}, http.StatusOK},
		{"regular developer is forbidden", &models.User{ID: "joe", IsAdmin: false}, http.StatusForbidden},
		{"anonymous is forbidden", nil, http.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(middlewares.NewSharedDataMiddleware())
			r.Use(injectPrincipal(tc.principal))
			r.Use(h.requireAdmin)
			r.Get("/admin", ok)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.want, rec.Code)
		})
	}
}

func TestAdminHandler_PostDeleteUser_CannotDeleteSelf(t *testing.T) {
	conf.Set(conf.Empty())
	admin := &models.User{ID: "root", IsAdmin: true}

	userMock := new(mocks.UserServiceMock)
	userMock.On("GetUserById", "root").Return(admin, nil)

	h := newAdminTestHandler(userMock)
	router := adminActionRouter(admin, h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/users/root/delete", nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)
	userMock.AssertNotCalled(t, "Delete", mock.Anything)
}

func TestAdminHandler_PostDeleteUser_DeletesOtherUser(t *testing.T) {
	conf.Set(conf.Empty())
	admin := &models.User{ID: "root", IsAdmin: true}
	target := &models.User{ID: "joe"}

	userMock := new(mocks.UserServiceMock)
	userMock.On("GetUserById", "joe").Return(target, nil)
	userMock.On("Delete", target).Return(nil)

	h := newAdminTestHandler(userMock)
	router := adminActionRouter(admin, h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/users/joe/delete", nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)
	userMock.AssertCalled(t, "Delete", target)
}

func TestAdminHandler_PostDeleteUser_NotFound(t *testing.T) {
	conf.Set(conf.Empty())
	admin := &models.User{ID: "root", IsAdmin: true}

	userMock := new(mocks.UserServiceMock)
	userMock.On("GetUserById", "ghost").Return(nil, errors.New("not found"))

	h := newAdminTestHandler(userMock)
	router := adminActionRouter(admin, h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/users/ghost/delete", nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)
	userMock.AssertNotCalled(t, "Delete", mock.Anything)
}

func TestAdminHandler_PostToggleAdmin_TogglesOtherUser(t *testing.T) {
	conf.Set(conf.Empty())
	admin := &models.User{ID: "root", IsAdmin: true}
	target := &models.User{ID: "joe", IsAdmin: false}

	userMock := new(mocks.UserServiceMock)
	userMock.On("GetUserById", "joe").Return(target, nil)
	userMock.On("Update", mock.MatchedBy(func(u *models.User) bool {
		return u.ID == "joe" && u.IsAdmin // promoted to admin
	})).Return(target, nil)

	h := newAdminTestHandler(userMock)
	router := adminActionRouter(admin, h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/users/joe/toggle-admin", nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)
	userMock.AssertCalled(t, "Update", mock.Anything)
}

func TestAdminHandler_PostCreateUser_Valid(t *testing.T) {
	conf.Set(conf.Empty())
	conf.Get().Mail.SkipVerifyMXRecord = true // don't do DNS MX lookups in tests
	admin := &models.User{ID: "root", IsAdmin: true}
	created := &models.User{ID: "newdev"}

	userMock := new(mocks.UserServiceMock)
	userMock.On("CreateOrGet", mock.MatchedBy(func(s *models.Signup) bool {
		return s.Username == "newdev" && s.Email == "new@example.test"
	}), false).Return(created, true, nil)

	h := newAdminTestHandler(userMock)
	router := adminActionRouter(admin, h)

	form := url.Values{"username": {"newdev"}, "email": {"new@example.test"}, "password": {"supersecret"}}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, postForm("/admin/users", form))

	assert.Equal(t, http.StatusFound, rec.Code)
	userMock.AssertCalled(t, "CreateOrGet", mock.Anything, false)
}

func TestAdminHandler_PostCreateUser_WeakPasswordRejected(t *testing.T) {
	conf.Set(conf.Empty())
	admin := &models.User{ID: "root", IsAdmin: true}

	userMock := new(mocks.UserServiceMock)
	h := newAdminTestHandler(userMock)
	router := adminActionRouter(admin, h)

	form := url.Values{"username": {"newdev"}, "password": {"short"}}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, postForm("/admin/users", form))

	assert.Equal(t, http.StatusFound, rec.Code)
	userMock.AssertNotCalled(t, "CreateOrGet", mock.Anything, mock.Anything)
}

func TestAdminHandler_PostEditUser(t *testing.T) {
	conf.Set(conf.Empty())
	conf.Get().Mail.SkipVerifyMXRecord = true // don't do DNS MX lookups in tests
	admin := &models.User{ID: "root", IsAdmin: true}
	target := &models.User{ID: "joe", Email: "old@example.test"}

	userMock := new(mocks.UserServiceMock)
	userMock.On("GetUserById", "joe").Return(target, nil)
	userMock.On("Update", mock.MatchedBy(func(u *models.User) bool {
		return u.ID == "joe" && u.Email == "new@example.test"
	})).Return(target, nil)

	h := newAdminTestHandler(userMock)
	router := adminActionRouter(admin, h)

	form := url.Values{"email": {"new@example.test"}}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, postForm("/admin/users/joe/edit", form))

	assert.Equal(t, http.StatusFound, rec.Code)
	userMock.AssertCalled(t, "Update", mock.Anything)
}

func TestAdminHandler_PostResetPassword(t *testing.T) {
	conf.Set(conf.Empty())
	admin := &models.User{ID: "root", IsAdmin: true}
	target := &models.User{ID: "joe"}

	userMock := new(mocks.UserServiceMock)
	userMock.On("GetUserById", "joe").Return(target, nil)
	userMock.On("Update", mock.MatchedBy(func(u *models.User) bool {
		return u.ID == "joe" && u.Password != "" // a hash was set
	})).Return(target, nil)

	h := newAdminTestHandler(userMock)
	router := adminActionRouter(admin, h)

	form := url.Values{"password": {"supersecret"}}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, postForm("/admin/users/joe/reset-password", form))

	assert.Equal(t, http.StatusFound, rec.Code)
	userMock.AssertCalled(t, "Update", mock.Anything)
}

func TestAdminHandler_BuildViewModel_SearchSortPaginate(t *testing.T) {
	conf.Set(conf.Empty())
	admin := &models.User{ID: "root", IsAdmin: true}
	users := []*models.User{
		{ID: "alice", Email: "alice@example.com"},
		{ID: "alex", Email: "alex@example.com"},
		{ID: "bob", Email: "bob@example.com"},
	}

	userMock := new(mocks.UserServiceMock)
	userMock.On("GetAll").Return(users, nil)
	userMock.On("CountCurrentlyOnline").Return(1, nil)

	hbMock := new(mocks.HeartbeatServiceMock)
	hbMock.On("CountByUsers", mock.Anything).Return([]*models.CountByUser{
		{User: "alice", Count: 9}, {User: "alex", Count: 3}, {User: "bob", Count: 1},
	}, nil)
	hbMock.On("GetLastAll").Return([]*models.TimeByUser{}, nil)
	hbMock.On("Count", true).Return(13, nil)

	h := &AdminHandler{config: conf.Get(), userService: userMock, heartbeatService: hbMock}

	var vm *view.AdminViewModel
	r := chi.NewRouter()
	r.Use(middlewares.NewSharedDataMiddleware())
	r.Use(injectPrincipal(admin))
	r.Get("/admin", func(w http.ResponseWriter, req *http.Request) {
		vm = h.buildViewModel(req, w)
	})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin?q=al", nil))

	assert.Equal(t, 3, vm.TotalUsers)
	assert.Equal(t, 2, vm.TotalMatched) // alice + alex match "al"
	require.Len(t, vm.Users, 2)
	// sorted by heartbeat count desc -> alice (9) before alex (3)
	assert.Equal(t, "alice", vm.Users[0].User.ID)
	assert.Equal(t, "alex", vm.Users[1].User.ID)
	assert.EqualValues(t, 13, vm.TotalHeartbeats)
}
