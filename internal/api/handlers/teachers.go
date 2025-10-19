package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
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

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	validField := map[string]bool{
		"name":    true,
		"email":   true,
		"class":   true,
		"subject": true,
	}

	return validField[field]
}

func getTeachersHandler(w http.ResponseWriter, r *http.Request, q *database.Queries) {
	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")

	if idStr == "" {
		var args []any

		sql := `
			SELECT id, name, email, class, subject, created_at, updated_at
			FROM teacher WHERE 1=1
		`

		params := map[string]string{
			"name":    "name",
			"email":   "email",
			"class":   "class",
			"subject": "subject",
		}

		paramIndex := 1
		for param, dbField := range params {
			value := r.URL.Query().Get(param)
			if value != "" {
				sql += fmt.Sprintf(" AND %s = $%d", dbField, paramIndex)
				args = append(args, value)
				paramIndex++
			}
		}

		sortParams := r.URL.Query()["sortby"]
		if len(sortParams) > 0 {
			sql += " ORDER BY"

			for i, param := range sortParams {
				parts := strings.Split(param, ":")
				if len(parts) != 2 {
					continue
				}

				field, order := parts[0], parts[1]

				if !isValidSortField(field) || !isValidSortOrder(order) {
					continue
				}

				if i > 0 {
					sql += ","
				}

				sql += fmt.Sprintf(" %s %s", field, order)
			}
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

type updateTeacherParams struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Class   string `json:"class"`
	Subject string `json:"subject"`
}

func putTeacherHandler(w http.ResponseWriter, r *http.Request, q *database.Queries) {
	var input updateTeacherParams

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}
	defer r.Body.Close()

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "The teacher ID must be a valid number.", http.StatusBadRequest)
		log.Printf("Invalid teacher ID '%s': %v", idStr, err)
		return
	}

	const sql = `
		UPDATE teacher SET
		name = $1,
		email = $2,
		class = $3,
		subject = $4
		WHERE id = $5
		RETURNING id, name, email, class, subject, created_at, updated_at
		`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var teacher models.Teacher
	row := q.DB.QueryRow(ctx, sql,
		&input.Name,
		&input.Email,
		&input.Class,
		&input.Subject,
		id,
	)

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
			http.Error(w, "Teacher not updated.", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to update teacher.", http.StatusInternalServerError)
		log.Printf("Database error during update: %v\n", err)
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

func patchTeacherHandler(w http.ResponseWriter, r *http.Request, q *database.Queries) {
	var input map[string]any

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v\n", err)
		return
	}
	defer r.Body.Close()

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "The teacher ID must be a valid number.", http.StatusBadRequest)
		log.Printf("Invalid teacher ID '%s': %v", idStr, err)
		return
	}

	const sqlSelect = `
		SELECT id, name, email, class, subject, created_at, updated_at FROM teacher WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := q.DB.QueryRow(ctx, sqlSelect, id)

	var teacher models.Teacher
	if err := row.Scan(
		&teacher.ID,
		&teacher.Name,
		&teacher.Email,
		&teacher.Class,
		&teacher.Subject,
		&teacher.CreatedAt,
		&teacher.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Teacher not found.", http.StatusNotFound)
			log.Printf("Teacher with ID %d not found.", id)
			return
		}
		log.Printf("Database query error (get teacher by ID): %v", err)
		http.Error(w, "Failed to fetch data from the database.", http.StatusInternalServerError)
		return
	}

	// for k, v := range input {
	// 	switch k {
	// 	case "name":
	// 		if name, ok := v.(string); ok {
	// 			teacher.Name = name
	// 		}
	// 	case "email":
	// 		if email, ok := v.(string); ok {
	// 			teacher.Email = email
	// 		}
	// 	case "class":
	// 		if class, ok := v.(string); ok {
	// 			teacher.Class = class
	// 		}
	// 	case "subject":
	// 		if subject, ok := v.(string); ok {
	// 			teacher.Subject = subject
	// 		}
	// 	}
	// }

	teacherValue := reflect.ValueOf(&teacher).Elem()
	teacherType := teacherValue.Type()

	for k, v := range input {
		for i := 0; i < teacherValue.NumField(); i++ {
			field := teacherType.Field(i)

			if field.Tag.Get("json") == k {
				if teacherValue.Field(i).CanSet() {
					teacherValue.Field(i).Set(reflect.ValueOf(v).Convert(teacherValue.Field(i).Type()))
				}
			}
		}
	}

	const sqlUpdate = `
		UPDATE teacher
		SET name = $1, email = $2, class = $3, subject = $4
		WHERE id = $5
	`

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = q.DB.Exec(ctx, sqlUpdate,
		&teacher.Name,
		&teacher.Email,
		&teacher.Class,
		&teacher.Subject,
		id,
	)
	if err != nil {
		log.Printf("Failed to update teacher with ID %d: %v", id, err)
		http.Error(w, "Failed to update teacher.", http.StatusInternalServerError)
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

func deleteTeacherHandler(w http.ResponseWriter, r *http.Request, q *database.Queries) {
	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Teacher ID must be a valid number.", http.StatusBadRequest)
		log.Printf("Invalid teacher ID format: %v", err)
		return
	}

	const sql = `
		DELETE FROM teacher WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := q.DB.Exec(ctx, sql, id)
	if err != nil {
		log.Printf("Error deleting teacher with ID %d: %v", id, err)
		http.Error(w, "Internal server error while deleting teacher.", http.StatusInternalServerError)
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("No teacher found with ID %d.", id),
		})
		log.Printf("No teacher found with ID %d", id)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": fmt.Sprintf("Teacher with ID %d successfully deleted.", id),
	}); err != nil {
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
		patchTeacherHandler(w, r, t.queries)

	case http.MethodPut:
		putTeacherHandler(w, r, t.queries)

	case http.MethodDelete:
		deleteTeacherHandler(w, r, t.queries)

	default:
		w.Write([]byte("Method Not Allowed"))
		fmt.Println("Method Not Allowed")
	}
}
