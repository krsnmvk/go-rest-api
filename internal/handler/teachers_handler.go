package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/krsnmvk/gorestapi/internal/database"
)

type TeachersHandler struct {
	q *database.Queries
}

func NewTeachersHandler(q *database.Queries) *TeachersHandler {
	return &TeachersHandler{
		q: q,
	}
}

type Teacher struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Class     string    `json:"class"`
	Subject   string    `json:"subject"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type createTeacherParams struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Class     string `json:"class"`
	Subject   string `json:"subject"`
}

func (th *TeachersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		pathParams := strings.TrimPrefix(r.URL.Path, "/teachers/")
		id := strings.TrimSuffix(pathParams, "/")
		fmt.Fprintf(w, "ID: %s\n", id)

		queryParams := r.URL.Query()
		name := queryParams.Get("name")
		fmt.Fprintf(w, "Name: %s\n", name)

		age := queryParams.Get("age")
		fmt.Fprintf(w, "Test: %s\n", age)

	case http.MethodPost:
		var input createTeacherParams

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			log.Printf("Error decoding JSON request body: %v\n", err)
			return
		}
		defer r.Body.Close()

		rctx := r.Context()

		const createQuery = `
			INSERT INTO teachers (
				first_name,
				last_name,
				email,
				class,
				subject
			) VALUES ($1, $2, $3, $4, $5) RETURNING 
				id, 
				first_name, 
				last_name, 
				email, 
				class, 
				subject, 
				created_at, 
				updated_at
			`

		var teacher Teacher
		row := th.q.Db.QueryRow(rctx, createQuery,
			&input.FirstName, &input.LastName, &input.Email, &input.Class, &input.Subject,
		)

		err := row.Scan(
			&teacher.ID,
			&teacher.FirstName,
			&teacher.LastName,
			&teacher.Email,
			&teacher.Class,
			&teacher.Subject,
			&teacher.CreatedAt,
			&teacher.UpdatedAt,
		)
		if err != nil {
			var pgErr *pgconn.PgError

			switch {
			case err == pgx.ErrNoRows:
				http.Error(w, "Teacher not found", http.StatusNotFound)

			case errors.As(err, &pgErr):
				log.Printf("Postgres error (%s): %s, Detail: %s, Hint: %s",
					pgErr.Code, pgErr.Message, pgErr.Detail, pgErr.Hint)
				http.Error(w, "Failed to create teacher", http.StatusInternalServerError)

			default:
				log.Printf("QueryRow Scan error: %v", err)
				http.Error(w, "Failed to create teacher", http.StatusInternalServerError)
			}
			return
		}

		resp := struct {
			Message string `json:"message"`
			Data    any    `json:"data"`
		}{
			Message: "Teacher data parsed successfully",
			Data:    teacher,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			fmt.Println("Error encoding response JSON:", err)
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}

	case http.MethodPut:
		w.Write([]byte("Hello PUT Method on Teachers Route"))
		fmt.Println("Hello PUT Method on Teachers Route")

	case http.MethodPatch:
		w.Write([]byte("Hello PATCH Method on Teachers Route"))
		fmt.Println("Hello PATCH Method on Teachers Route")

	case http.MethodDelete:
		w.Write([]byte("Hello DELETE Method on Teachers Route"))
		fmt.Println("Hello DELETE Method on Teachers Route")

	default:
		w.Write([]byte("Method Not Allowed!"))
		fmt.Println("Method Not Allowed!")
	}
}
