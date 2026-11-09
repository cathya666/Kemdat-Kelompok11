package middlewares

import (
	"aaa_bekend/internal/lmodel"
	"net/http"
)

type RoleGuard struct {
	next         http.Handler
	allowedRoles []string
}

func (rg *RoleGuard) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	user, err := lmodel.UserFromContext(r.Context())
	// TODO

	http.Error(w, "Forbidden", http.StatusForbidden)

}

func RequireRole(allowedRules ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &RoleGuard{
			next:         next,
			allowedRoles: allowedRules,
		}
	}
}
