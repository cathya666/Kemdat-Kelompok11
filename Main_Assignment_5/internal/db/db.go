package db

import (
	"aaa_bekend/internal/lmodel"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Init() {
	ctx := context.Background()
	dsn := "postgres://security_user:super_secret_password@localhost:5432/aaa_exhibition?sslmode=disable"
	var err error

	Pool, err = pgxpool.New(ctx, dsn)
	if err != nil {

		log.Fatalf("Unable to connect to database: %v\n", err)
		return
	}
	err = Pool.Ping(ctx)

	if err != nil {
		log.Fatalf("Unable to ping the database: %v\n", err)
		return
	}

	fmt.Println("Successfully connected to PostGreSQL")
}

// <Memes>

func CreateUser(ctx context.Context, username, password_hash, role string) error {
	query := `INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3)`
	_, err := Pool.Exec(ctx, query, username, password_hash, role)
	if err != nil {
		return err
	}
	return nil
}

func GetUserByUsername(ctx context.Context, username string) (*lmodel.User, error) {
	var user *lmodel.User = &lmodel.User{}
	query := `SELECT id, username, password_hash, role FROM USERS where username=$1`
	err := Pool.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CreateProject(ctx context.Context, title string, description string, student_id int) (int, error) {
	var projectID int
	query := `INSERT INTO projects (title, description, student_id) VALUES ($1, $2, $3) RETURNING id`
	err := Pool.QueryRow(ctx, query, title, description, student_id).Scan(&projectID)
	if err != nil {
		return -1, err
	}
	return projectID, nil
}

func GetProjectByID(ctx context.Context, id int) (*lmodel.Project, error) {
	p := &lmodel.Project{}
	query := `SELECT id, title, description, student_id FROM projects WHERE id = $1`
	err := Pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.Title, &p.Description, &p.StudentID)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func GetAllProjects(ctx context.Context) ([]lmodel.Project, error) {
	var projects []lmodel.Project
	var query string = `SELECT id, title, description, student_id from projects`

	rows, err := Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p lmodel.Project
		err = rows.Scan(&p.ID, &p.Title, &p.Description, &p.StudentID)
		if err != nil {
			return nil, err
		}

		projects = append(projects, p)

	}

	return projects, nil
}

func GetAllProjectGrades(ctx context.Context) ([]lmodel.ProjectGrade, error) {
	var query string = `SELECT 
			p.id, p.title, p.description, p.student_id,
			g.id, g.project_id, g.judge_id, g.score, g.feedback
			FROM projects p 
			INNER JOIN grades g ON p.id=g.project_id
			ORDER BY p.id, g.id`
	rows, err := Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var projectGrades []lmodel.ProjectGrade
	for rows.Next() {
		var pg lmodel.ProjectGrade
		err := rows.Scan(
			&pg.Proj.ID, &pg.Proj.Title, &pg.Proj.Description, &pg.Proj.StudentID,
			&pg.Grad.ID, &pg.Grad.ProjectID, &pg.Grad.JudgeID, &pg.Grad.Score, &pg.Grad.Feedback,
		)
		if err != nil {
			return nil, err
		}
		projectGrades = append(projectGrades, pg)
	}

	return projectGrades, nil

}

func CreateGrade(ctx context.Context, projectID int, judgeId int, score int, feedback string) error {
	var query string = `INSERT INTO grades (project_id, judge_id, score, feedback) VALUES ($1, $2, $3, $4)
	ON CONFLICT (project_id, judge_id)
	DO UPDATE SET score = EXCLUDED.score, feedback = EXCLUDED.feedback`
	_, err := Pool.Exec(ctx, query, projectID, judgeId, score, feedback)
	if err != nil {
		return err
	}
	return nil
}

func GetGradesForProject(ctx context.Context, projectID int) ([]lmodel.Grade, error) {
	var grades []lmodel.Grade
	var query string = `SELECT id, project_id, judge_id, score, feedback from grades where project_id = $1`
	rows, err := Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var g lmodel.Grade
		err = rows.Scan(&g.ID, &g.ProjectID, &g.JudgeID, &g.Score, &g.Feedback)
		if err != nil {
			return nil, err
		}

		grades = append(grades, g)
	}
	return grades, nil
}

func InsertAuditLog(ctx context.Context, auditLog lmodel.AuditLog) {
	query := `INSERT INTO audit_logs (user_id, username, method, path, status_code, ip_address, user_agent) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`
	go func() {
		_, err := Pool.Exec(context.Background(), query,
			auditLog.UserID, auditLog.Username, auditLog.Method,
			auditLog.Path, auditLog.StatusCode, auditLog.IPAddress,
			auditLog.UserAgent)
		if err != nil {
			log.Printf("Failed to insert audit log: %v", err)
		}
	}()
}

// / Memes
func GetUserByID(ctx context.Context, userID int) (*lmodel.User, error) {
	var user *lmodel.User = &lmodel.User{}
	query := "SELECT id, username, role, created_at, updated_at FROM users WHERE id=$1"
	err := Pool.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UpdateMyData(ctx context.Context, userID int, username string) error {
	query := `UPDATE users SET username = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := Pool.Exec(ctx, query, username, userID)
	if err != nil {
		return err
	}
	return nil
}

func GetUserAuditLogs(ctx context.Context, userID int) ([]lmodel.AuditLog, error) {
	query := `SELECT method, path, status_code, ip_address, timestamp FROM 
			audit_logs WHERE user_id=$1 
			ORDER BY timestamp DESC LIMIT 50
		`
	rows, err := Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []lmodel.AuditLog

	for rows.Next() {
		var log lmodel.AuditLog
		rows.Scan(&log.Method, &log.Path, &log.StatusCode, &log.IPAddress, &log.Timestamp)
		logs = append(logs, log)
	}
	return logs, nil
}

func GetUserProjects(ctx context.Context, userID int) ([]lmodel.Project, error) {

	query := "SELECT id, title, description, student_id FROM projects WHERE student_id=$1"

	rows, err := Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var projects []lmodel.Project

	for rows.Next() {
		var project lmodel.Project
		err = rows.Scan(&project.ID, &project.Title, &project.Description, &project.StudentID)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func DeleteUserAccount(ctx context.Context, userID int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := Pool.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}
