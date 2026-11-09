package main

import (
	"aaa_bekend/internal/auth"
	"aaa_bekend/internal/db"
	"aaa_bekend/internal/handler"
	"aaa_bekend/internal/middlewares"
	"fmt"
	"log"
	"net/http"
)

func main() {
	db.Init()
	defer db.Pool.Close()

	err := auth.LoadKeys()
	if err != nil {
		log.Fatalf("failed to load RSA keys. Please run make keys first")
	}
	mux := http.NewServeMux()

	mux.Handle("POST /api/v1/register", (http.HandlerFunc(handler.Register)))
	mux.Handle("POST /api/v1/login", http.HandlerFunc(handler.Login))

	mux.Handle("GET /api/v1/projects", middlewares.AuthN(middlewares.Audit(http.HandlerFunc(handler.ListProjects))))
	mux.Handle("POST /api/v1/projects",
		middlewares.AuthN(
			middlewares.Audit(
				middlewares.RequireRole("student")(
					http.HandlerFunc(handler.SubmitProject),
				),
			),
		),
	)

	mux.Handle("POST /api/v1/projects/{id}/grade",
		middlewares.AuthN(
			middlewares.RequireRole("judge", "admin")(
				middlewares.Audit(
					http.HandlerFunc(handler.GradeProject),
				),
			),
		),
	)

	mux.Handle("GET /api/v1/projects/grade",
		middlewares.AuthN(
			middlewares.RequireRole("judge", "admin")(
				middlewares.Audit(
					http.HandlerFunc(handler.GetAllProjectGrades),
				),
			),
		),
	)

	mux.Handle("GET /api/v1/projects/{id}/grades",
		middlewares.AuthN(
			middlewares.Audit(
				http.HandlerFunc(handler.GetProjectGrades),
			),
		),
	)

	// Exclusively for UU PDP
	mux.Handle("GET /api/v1/users/me/data",
		middlewares.AuthN(
			middlewares.Audit(
				http.HandlerFunc(handler.GetMyData),
			),
		),
	)
	mux.Handle("PUT /api/v1/users/me/data",
		middlewares.AuthN(
			middlewares.Audit(
				http.HandlerFunc(handler.UpdateMyData),
			),
		),
	)
	mux.Handle("GET /api/v1/users/me/export",
		middlewares.AuthN(
			middlewares.Audit(
				http.HandlerFunc(handler.ExportMyData),
			),
		),
	)
	mux.Handle("GET /api/v1/users/me/audit-log",
		middlewares.AuthN(
			middlewares.Audit(
				http.HandlerFunc(handler.GetMyAuditLog),
			),
		),
	)

	mux.Handle("DELETE /api/v1/users/me",
		middlewares.AuthN(
			middlewares.Audit(
				http.HandlerFunc(handler.DeleteMyAccount),
			),
		),
	)

	fmt.Println("🚀 Exhibition Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
