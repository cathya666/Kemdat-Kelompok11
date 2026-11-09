package main

import (
	"aaa_bekend/internal/auth"
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/xuri/excelize/v2"
)

const (
	dsn      = "postgres://security_user:super_secret_password@localhost:5432/aaa_exhibition?sslmode=disable" // Change to your DSN
	seedFile = "seeds/seed_data.xlsx"
)

func main() {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}
	defer conn.Close(ctx)
	fmt.Println("Connected to database")

	file, err := excelize.OpenFile(seedFile)
	if err != nil {
		log.Fatalf("Cannot open %s: %v", seedFile, err)
	}
	defer file.Close()
	fmt.Printf("Opened %s\n", seedFile)

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("Cannot begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	userIDs := seedUsers(ctx, tx, file)
	projectIDs := seedProjects(ctx, tx, file, userIDs)
	seedGrades(ctx, tx, file, userIDs, projectIDs)

	// 4. Commit
	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("Cannot commit: %v", err)
	}
	fmt.Println("\nDatabase seeded successfully using Argon2id!")
}

func seedUsers(ctx context.Context, tx pgx.Tx, file *excelize.File) map[string]int {
	rows := readSheet(file, "Users")
	userIDs := make(map[string]int)

	for _, row := range rows {
		username := row[0]
		plainPassword := row[1]
		role := row[2]

		hashedPassword, err := auth.HashPassword(plainPassword)
		if err != nil {
			log.Fatalf("Cannot hash password for %s: %v", username, err)
		}

		var id int
		err = tx.QueryRow(ctx,
			`INSERT INTO users (username, password_hash, role)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (username) DO UPDATE SET role = EXCLUDED.role, password_hash = EXCLUDED.password_hash
			 RETURNING id`,
			username, hashedPassword, role,
		).Scan(&id)
		if err != nil {
			log.Fatalf("Cannot insert user %s: %v", username, err)
		}

		userIDs[username] = id
		fmt.Printf("user: %-16s (%s)\n", username, role)
	}
	return userIDs
}

// ─────────────────────────────────────────────
func seedProjects(ctx context.Context, tx pgx.Tx, file *excelize.File, userIDs map[string]int) map[string]int {
	rows := readSheet(file, "Projects")
	projectIDs := make(map[string]int)

	for _, row := range rows {
		title := row[0]
		description := row[1]
		studentUsername := row[2]

		studentID, ok := userIDs[studentUsername]
		if !ok {
			log.Fatalf("Student %s not found in Users sheet", studentUsername)
		}

		// Skip if this project already exists (idempotent)
		var id int
		err := tx.QueryRow(ctx,
			`SELECT id FROM projects WHERE title = $1 AND student_id = $2`,
			title, studentID,
		).Scan(&id)

		if err == pgx.ErrNoRows {
			// Not found → insert it
			err = tx.QueryRow(ctx,
				`INSERT INTO projects (title, description, student_id)
				 VALUES ($1, $2, $3) RETURNING id`,
				title, description, studentID,
			).Scan(&id)
			if err != nil {
				log.Fatalf("Cannot insert project %s: %v", title, err)
			}
			fmt.Printf("project: %s\n", title)
		}

		projectIDs[title] = id
	}
	return projectIDs
}

func seedGrades(ctx context.Context, tx pgx.Tx, file *excelize.File, userIDs map[string]int, projectIDs map[string]int) {
	rows := readSheet(file, "Grades")

	for _, row := range rows {
		projectTitle := row[0]
		judgeUsername := row[1]
		score, _ := strconv.Atoi(row[2])
		feedback := row[3]

		projectID, ok := projectIDs[projectTitle]
		if !ok {
			log.Fatalf("Project %s not found in Projects sheet", projectTitle)
		}
		judgeID, ok := userIDs[judgeUsername]
		if !ok {
			log.Fatalf("Judge %s not found in Users sheet", judgeUsername)
		}

		_, err := tx.Exec(ctx,
			`INSERT INTO grades (project_id, judge_id, score, feedback)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (project_id, judge_id) DO NOTHING`,
			projectID, judgeID, score, feedback,
		)
		if err != nil {
			log.Fatalf("Cannot insert grade for %s: %v", projectTitle, err)
		}
		fmt.Printf("grade: %s → %s (%d)\n", judgeUsername, projectTitle, score)
	}
}

func readSheet(file *excelize.File, sheetName string) [][]string {
	rows, err := file.GetRows(sheetName)
	if err != nil {
		log.Fatalf("Cannot read sheet %q: %v", sheetName, err)
	}

	var data [][]string
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		if len(row) == 0 || row[0] == "" {
			continue // skip empty rows
		}
		data = append(data, row)
	}
	return data
}
