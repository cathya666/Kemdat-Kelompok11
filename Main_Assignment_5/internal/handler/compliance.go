package handler

import (
	"aaa_bekend/internal/db"
	"aaa_bekend/internal/lmodel"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ---------- helpers ----------

func complianceJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func complianceError(w http.ResponseWriter, status int, msg string) {
	complianceJSON(w, status, map[string]string{"error": msg})
}

// ---------- handlers ----------

func GetMyData(w http.ResponseWriter, r *http.Request) {
	user, err := lmodel.UserFromContext(r.Context())
	if err != nil {
		complianceError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	profile, err := db.GetUserByID(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			complianceError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("GetMyData: %v", err)
		complianceError(w, http.StatusInternalServerError, "Failed to retrieve user data")
		return
	}

	complianceJSON(w, http.StatusOK, profile)
}

func UpdateMyData(w http.ResponseWriter, r *http.Request) {
	user, err := lmodel.UserFromContext(r.Context())
	if err != nil {
		complianceError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // batasi body 1 MB
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		complianceError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		complianceError(w, http.StatusBadRequest, "Username must not be empty")
		return
	}
	if utf8.RuneCountInString(req.Username) > 50 { // sesuai VARCHAR(50)
		complianceError(w, http.StatusBadRequest, "Username must be at most 50 characters")
		return
	}

	if err := db.UpdateMyData(r.Context(), user.ID, req.Username); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			complianceError(w, http.StatusConflict, "Username already taken")
			return
		}
		log.Printf("UpdateMyData: %v", err)
		complianceError(w, http.StatusInternalServerError, "Failed to update data")
		return
	}

	complianceJSON(w, http.StatusOK, map[string]string{
		"message":  "Data updated successfully",
		"username": req.Username,
	})
}

func GetMyAuditLog(w http.ResponseWriter, r *http.Request) {
	user, err := lmodel.UserFromContext(r.Context())
	if err != nil {
		complianceError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	logs, err := db.GetUserAuditLogs(r.Context(), user.ID)
	if err != nil {
		log.Printf("GetMyAuditLog: %v", err)
		complianceError(w, http.StatusInternalServerError, "Failed to retrieve audit logs")
		return
	}
	if logs == nil {
		logs = []lmodel.AuditLog{} // supaya JSON berisi [] bukan null
	}

	complianceJSON(w, http.StatusOK, map[string]interface{}{
		"user":  user.Username,
		"total": len(logs),
		"logs":  logs,
	})
}

func ExportMyData(w http.ResponseWriter, r *http.Request) {
	user, err := lmodel.UserFromContext(r.Context())
	if err != nil {
		complianceError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	profile, err := db.GetUserByID(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			complianceError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("ExportMyData (profile): %v", err)
		complianceError(w, http.StatusInternalServerError, "Failed to export data")
		return
	}

	projects, err := db.GetUserProjects(r.Context(), user.ID)
	if err != nil {
		log.Printf("ExportMyData (projects): %v", err)
		complianceError(w, http.StatusInternalServerError, "Failed to export data")
		return
	}
	if projects == nil {
		projects = []lmodel.Project{}
	}

	logs, err := db.GetUserAuditLogs(r.Context(), user.ID)
	if err != nil {
		log.Printf("ExportMyData (audit logs): %v", err)
		complianceError(w, http.StatusInternalServerError, "Failed to export data")
		return
	}
	if logs == nil {
		logs = []lmodel.AuditLog{}
	}

	export := map[string]interface{}{
		"exported_at": time.Now(),
		"profile":     profile,
		"projects":    projects,
		"audit_logs":  logs,
	}

	w.Header().Set("Content-Disposition", "attachment; filename=my_data_export.json")
	complianceJSON(w, http.StatusOK, export)
}

func DeleteMyAccount(w http.ResponseWriter, r *http.Request) {
	user, err := lmodel.UserFromContext(r.Context())
	if err != nil {
		complianceError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Langkah 1: hapus audit log milik user lebih dulu
	// (FK audit_logs.user_id = ON DELETE SET NULL, jadi tidak ikut terhapus otomatis).
	if _, err := db.Pool.Exec(r.Context(), `DELETE FROM audit_logs WHERE user_id = $1`, user.ID); err != nil {
		log.Printf("DeleteMyAccount (audit logs): %v", err)
		complianceError(w, http.StatusInternalServerError, "Failed to delete audit logs")
		return
	}

	// Langkah 2: hapus akun. Projects & grades terhapus lewat ON DELETE CASCADE.
	if err := db.DeleteUserAccount(r.Context(), user.ID); err != nil {
		log.Printf("DeleteMyAccount (user): %v", err)
		complianceError(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	complianceJSON(w, http.StatusOK, map[string]string{
		"message": "Your account and all data have been permanently deleted",
	})
}
