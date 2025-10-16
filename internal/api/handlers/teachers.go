package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/krsnmvk/gorestapi/internal/database"
	"github.com/krsnmvk/gorestapi/internal/models"
)

type Teacher struct {
	queries *database.Queries
}

func NewTeacher(queries *database.Queries) *Teacher {
	return &Teacher{queries: queries}
}

type createTeacherParams struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Class   string `json:"class"`
	Subject string `json:"subject"`
}

type response[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

func getTeachersHandler(w http.ResponseWriter, r *http.Request, q *database.Queries) {
	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")

	if idStr == "" {
		name := r.URL.Query().Get("name")
		var args []any

		sql := `
			SELECT id, name, email, class, subject, created_at, updated_at
			FROM teacher WHERE 1=1
		`

		if name != "" {
			sql += " AND name = $1 LIMIT 10"
			args = append(args, name)
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		rows, err := q.DB.Query(ctx, sql, args...)
		if err != nil {
			log.Printf("Database query error (teacher list): %v", err)
			http.Error(w, "Failed to fetch data from the database.", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var teachers []models.Teacher
		for rows.Next() {
			var t models.Teacher
			err := rows.Scan(
				&t.ID,
				&t.Name,
				&t.Email,
				&t.Class,
				&t.Subject,
				&t.CreatedAt,
				&t.UpdatedAt,
			)
			if err != nil {
				log.Printf("Failed to scan teacher row: %v", err)
				http.Error(w, "Failed to process teacher data.", http.StatusInternalServerError)
				return
			}
			teachers = append(teachers, t)
		}

		if err := rows.Err(); err != nil {
			log.Printf("Row iteration error: %v", err)
			http.Error(w, "Unexpected error occurred while reading teachers.", http.StatusInternalServerError)
			return
		}

		resp := response[[]models.Teacher]{
			Success: true,
			Data:    teachers,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
			http.Error(w, "Unable to send response.", http.StatusInternalServerError)
		}
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "The teacher ID must be a valid number.", http.StatusBadRequest)
		log.Printf("Invalid teacher ID '%s': %v", idStr, err)
		return
	}

	const sql = `
		SELECT id, name, email, class, subject, created_at, updated_at
		FROM teacher WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var teacher models.Teacher
	row := q.DB.QueryRow(ctx, sql, id)
	err = row.Scan(
		&teacher.ID,
		&teacher.Name,
		&teacher.Email,
		&teacher.Class,
		&teacher.Subject,
		&teacher.CreatedAt,
		&teacher.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Teacher not found.", http.StatusNotFound)
			log.Printf("Teacher with ID %d not found.", id)
			return
		}

		log.Printf("Database query error (get teacher by ID): %v", err)
		http.Error(w, "Failed to fetch data from the database.", http.StatusInternalServerError)
		return
	}

	resp := response[models.Teacher]{
		Success: true,
		Data:    teacher,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
		http.Error(w, "Unable to send response.", http.StatusInternalServerError)
	}
}

func postTeacherHandler(w http.ResponseWriter, r *http.Request, q *database.Queries) {
	var input createTeacherParams

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}
	defer r.Body.Close()

	const sql = `
		INSERT INTO teacher (name, email, class, subject)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, class, subject, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var teacher models.Teacher
	row := q.DB.QueryRow(ctx, sql, input.Name, input.Email, input.Class, input.Subject)
	err := row.Scan(
		&teacher.ID,
		&teacher.Name,
		&teacher.Email,
		&teacher.Class,
		&teacher.Subject,
		&teacher.CreatedAt,
		&teacher.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				http.Error(w, "A teacher with this email already exists.", http.StatusConflict)
				log.Printf("Unique constraint violation (email): %v", pgErr)
				return
			}
		}

		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Teacher not created.", http.StatusInternalServerError)
			log.Printf("Insert returned no rows.")
			return
		}

		http.Error(w, "Failed to create teacher.", http.StatusInternalServerError)
		log.Printf("Database insert error: %v", err)
		return
	}

	resp := response[models.Teacher]{
		Success: true,
		Data:    teacher,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
		http.Error(w, "Unable to send response.", http.StatusInternalServerError)
	}
}

func (t *Teacher) TeachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTeachersHandler(w, r, t.queries)

	case http.MethodPost:
		postTeacherHandler(w, r, t.queries)

	case http.MethodPatch:
		w.Write([]byte("Hello PATCH Method on Teachers Route"))
		fmt.Println("Hello PATCH Method on Teachers Route")

	case http.MethodPut:
		w.Write([]byte("Hello PUT Method on Teachers Route"))
		fmt.Println("Hello PUT Method on Teachers Route")

	case http.MethodDelete:
		w.Write([]byte("Hello DELETE Method on Teachers Route"))
		fmt.Println("Hello DELETE Method on Teachers Route")

	default:
		w.Write([]byte("Method Not Allowed"))
		fmt.Println("Method Not Allowed")
	}
}
