package lmodel

import (
	"context"
	"errors"
	"time"
)

type User struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type Project struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StudentID   int    `json:"student_id"`
}

type Grade struct {
	ID        int    `json:"id"`
	ProjectID int    `json:"project_id"`
	JudgeID   int    `json:"judge_id"`
	Score     int    `json:"score"`
	Feedback  string `json:"feedback"`
}

type ProjectGrade struct {
	Proj Project `json:"project"`
	Grad Grade   `json:"grade"`
}

type AuditLog struct {
	ID         int
	UserID     int
	Username   string
	Method     string
	Path       string
	StatusCode int
	IPAddress  string
	UserAgent  string
	Timestamp  time.Time
}

type DataExport struct {
	User       User       `json:"user_profile"`
	Projects   []Project  `json:"projects,omitempty"`
	Grades     []Grade    `json:"grades_given,omitempty"`
	AuditLog   []AuditLog `json:"audit_logs,omitempty"`
	ExportedAt time.Time  `json:"exported_at"`
}

type contextKey string

const userContextKey contextKey = "user"

type UserContext struct {
	ID       int
	Username string
	Role     string
}

func ContextWithUser(ctx context.Context, user UserContext) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (UserContext, error) {
	user, ok := ctx.Value(userContextKey).(UserContext)
	if !ok {
		return UserContext{}, errors.New("user not found in context")
	}
	return user, nil
}
