package middlewares

import (
	"aaa_bekend/internal/db"
	"aaa_bekend/internal/lmodel"
	"log"
	"net"
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Audit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		var userCtx lmodel.UserContext
		if ctxUser, err := lmodel.UserFromContext(r.Context()); err == nil {
			userCtx = ctxUser
		}

		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		} else if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}

		auditLog := lmodel.AuditLog{
			UserID:     userCtx.ID,
			Username:   userCtx.Username,
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: rw.statusCode,
			IPAddress:  ip,
			UserAgent:  r.UserAgent(),
		}

		db.InsertAuditLog(r.Context(), auditLog)
		log.Printf("[AUDIT] %s %s | User: %s | Status: %d | IP: %s",
			auditLog.Method, auditLog.Path, auditLog.Username, auditLog.StatusCode, auditLog.IPAddress)
	})
}
