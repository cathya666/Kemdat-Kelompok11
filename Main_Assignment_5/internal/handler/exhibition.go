package handler

import (
	"aaa_bekend/internal/db"
	"aaa_bekend/internal/lmodel"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type SubmitProjectReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GradeProjectReq struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

// POST /api/v1/projects
func SubmitProject(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := lmodel.UserFromContext(r.Context())

	// TODO 1: Decode request body
	var req SubmitProjectReq

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO 2: Simpan proyek ke database
	projectID, err := db.CreateProject(
		r.Context(),
		req.Title,
		req.Description,
		userCtx.ID,
	)

	if err != nil {
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{
		"project_id": projectID,
	})
}

func ListProjects(w http.ResponseWriter, r *http.Request) {
	// TODO 3: Ambil semua proyek
	projects, err := db.GetAllProjects(r.Context())

	if err != nil {
		http.Error(w, "Failed to fetch projects", http.StatusInternalServerError)
		return
	}

	// TODO 4: Antisipasi data kosong
	if projects == nil {
		projects = []lmodel.Project{}
	}

	json.NewEncoder(w).Encode(&projects)
}

func GetAllProjectGrades(w http.ResponseWriter, r *http.Request) {
	// TODO 5: Ambil seluruh proyek beserta nilainya
	projects, err := db.GetAllProjectGrades(r.Context())

	if err != nil {
		http.Error(w, "Failed to fetch project grades", http.StatusInternalServerError)
		return
	}

	// TODO 6: Antisipasi data kosong
	if projects == nil {
		projects = []lmodel.ProjectGrade{}
	}

	json.NewEncoder(w).Encode(&projects)
}

func GradeProject(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := lmodel.UserFromContext(r.Context())

	// TODO 7: Ambil ID proyek dari URL
	projectIDStr := r.PathValue("id")

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// TODO 8: Decode request body
	var req GradeProjectReq

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO 9: Simpan nilai ke database
	err = db.CreateGrade(
		r.Context(),
		projectID,
		userCtx.ID,
		req.Score,
		req.Feedback,
	)

	if err != nil {
		http.Error(w, "Failed to save grade", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Grade Submitted",
	})
}

func GetProjectGrades(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := lmodel.UserFromContext(r.Context())

	// TODO 10: Ambil ID proyek dari URL
	projectIDStr := r.PathValue("id")

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// TODO 11: Ambil detail proyek dari database
	project, err := db.GetProjectByID(r.Context(), projectID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to fetch project", http.StatusInternalServerError)
		return
	}

	// Authorization: Student hanya dapat melihat nilai proyek miliknya
	if userCtx.Role == "student" {
		if project.StudentID != userCtx.ID {
			http.Error(
				w,
				"Forbidden: you can only view grades for your own projects",
				http.StatusForbidden,
			)
			return
		}
	}

	// Judge dan Admin dapat melihat nilai proyek mana pun
	grades, err := db.GetGradesForProject(r.Context(), projectID)

	if err != nil {
		http.Error(w, "Failed to fetch grades", http.StatusInternalServerError)
		return
	}

	if grades == nil {
		grades = []lmodel.Grade{}
	}

	json.NewEncoder(w).Encode(grades)
}
