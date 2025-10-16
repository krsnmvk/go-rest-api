package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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

func postTeacherHandler(w http.ResponseWriter, r *http.Request, q *database.Queries) {
	var input createTeacherParams

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	const sql = `
	INSERT INTO teacher (name, email, class, subject)
	VALUES ($1, $2, $3, $4)
	RETURNING id, name, email, class, subject, created_at, updated_at
`

	var teacher models.Teacher

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := q.DB.QueryRow(ctx, sql, &input.Name, &input.Email, &input.Class, &input.Subject)

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
				http.Error(w, "Email already in use.", http.StatusConflict)
				return
			}

		}

		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "No rows found.", http.StatusNotFound)
			return
		}

		http.Error(w, "Query failed.", http.StatusInternalServerError)
		log.Printf("Query failed: %v\n", err)
		return
	}

	resp := response[models.Teacher]{
		Success: true,
		Data:    teacher,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode JSON response.", http.StatusInternalServerError)
		log.Printf("An error ocurred while sending the response: %v\n", err)
		return
	}
}

func (t *Teacher) TeachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello GET Method on Teachers Route"))
		fmt.Println("Hello GET Method on Teachers Route")

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
