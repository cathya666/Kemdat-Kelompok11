package middlewares

import (
	"aaa_bekend/internal/lmodel"
	"net/http"
)

func AuthN(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// TODO

		// TODO

		ctx := lmodel.ContextWithUser(r.Context(), userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
