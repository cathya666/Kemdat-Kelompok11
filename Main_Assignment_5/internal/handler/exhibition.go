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
	// TODO

	// TODO
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"project_id": projectID})
}

func ListProjects(w http.ResponseWriter, r *http.Request) {
	// TODO

	// TODO

	json.NewEncoder(w).Encode(&projects)
}
func GetAllProjectGrades(w http.ResponseWriter, r *http.Request) {
	// TODO

	// TODO

	json.NewEncoder(w).Encode(&projects)
}
func GradeProject(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := lmodel.UserFromContext(r.Context())

	//TODO

	// TODO
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Grade Submitted"})
}

func GetProjectGrades(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := lmodel.UserFromContext(r.Context())
	
	// TODO 

	// TODO

		if project.StudentID != userCtx.ID {
			http.Error(w, "Forbidden: you can only view grades for your own projects", http.StatusForbidden)
			return
		}

	}

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
