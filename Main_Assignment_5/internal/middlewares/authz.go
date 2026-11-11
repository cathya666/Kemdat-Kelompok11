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

	// memastikan user sudah terautentikasi
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// memeriksa apakah role user termasuk role yang diizinkan
	for _, role := range rg.allowedRoles {
		if user.Role == role {
			rg.next.ServeHTTP(w, r)
			return
		}
	}

	// Role tidak memiliki izin
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
