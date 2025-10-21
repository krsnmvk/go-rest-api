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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krsnmvk/gorestapi/internal/database"
	"github.com/krsnmvk/gorestapi/internal/models"
)

type TeacherHandler struct {
	db   *database.Queries
	pool *pgxpool.Pool
}

func NewTeacherHandler(db *database.Queries, pool *pgxpool.Pool) *TeacherHandler {
	return &TeacherHandler{db: db, pool: pool}
}

type CreateTeacherParams struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Class   string `json:"class"`
	Subject string `json:"subject"`
}

type APIResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

// ==========================================
// GET /teachers/{id}
// ==========================================
func (h *TeacherHandler) GetTeacher(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Teacher ID must be a valid number.", http.StatusBadRequest)
		log.Printf("Invalid teacher ID '%s': %v", idStr, err)
		return
	}

	const query = `
		SELECT id, name, email, class, subject, created_at, updated_at
		FROM teacher WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var teacher models.Teacher
	err = h.db.DB.QueryRow(ctx, query, id).Scan(
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

		log.Printf("Database query error (GetTeacher): %v", err)
		http.Error(w, "Failed to fetch data from the database.", http.StatusInternalServerError)
		return
	}

	response := APIResponse[models.Teacher]{Success: true, Data: teacher}
	writeJSON(w, http.StatusOK, response)
}

// ==========================================
// POST /teachers
// ==========================================
func (h *TeacherHandler) CreateTeacher(w http.ResponseWriter, r *http.Request) {
	var input CreateTeacherParams

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}
	defer r.Body.Close()

	const query = `
		INSERT INTO teacher (name, email, class, subject)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, class, subject, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var teacher models.Teacher
	err := h.db.DB.QueryRow(ctx, query, input.Name, input.Email, input.Class, input.Subject).Scan(
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
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "A teacher with this email already exists.", http.StatusConflict)
			log.Printf("Unique constraint violation (email): %v", pgErr)
			return
		}

		http.Error(w, "Failed to create teacher.", http.StatusInternalServerError)
		log.Printf("Database insert error: %v", err)
		return
	}

	response := APIResponse[models.Teacher]{Success: true, Data: teacher}
	writeJSON(w, http.StatusCreated, response)
}

// ==========================================
// GET /teachers
// ==========================================
func (h *TeacherHandler) GetTeachers(w http.ResponseWriter, r *http.Request) {
	var args []any

	query := `
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
		if value := r.URL.Query().Get(param); value != "" {
			query += fmt.Sprintf(" AND %s = $%d", dbField, paramIndex)
			args = append(args, value)
			paramIndex++
		}
	}

	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		query += " ORDER BY"

		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}

			field, order := parts[0], strings.ToLower(parts[1])
			if !isValidSortField(field) || !isValidSortOrder(order) {
				continue
			}

			if i > 0 {
				query += ","
			}

			query += fmt.Sprintf(" %s %s", field, order)
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := h.db.DB.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Database query error (GetTeachers): %v", err)
		http.Error(w, "Failed to fetch data from the database.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var teachers []models.Teacher
	for rows.Next() {
		var teacher models.Teacher
		if err := rows.Scan(
			&teacher.ID,
			&teacher.Name,
			&teacher.Email,
			&teacher.Class,
			&teacher.Subject,
			&teacher.CreatedAt,
			&teacher.UpdatedAt,
		); err != nil {
			log.Printf("Failed to scan teacher row: %v", err)
			http.Error(w, "Failed to process teacher data.", http.StatusInternalServerError)
			return
		}
		teachers = append(teachers, teacher)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		http.Error(w, "Unexpected error occurred while reading teachers.", http.StatusInternalServerError)
		return
	}

	response := APIResponse[[]models.Teacher]{Success: true, Data: teachers}
	writeJSON(w, http.StatusOK, response)
}

// ==========================================
// PUT /teachers/{id}
// ==========================================
func (h *TeacherHandler) UpdateTeacher(w http.ResponseWriter, r *http.Request) {
	var input map[string]any

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}
	defer r.Body.Close()

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Teacher ID must be a valid number.", http.StatusBadRequest)
		return
	}

	const selectQuery = `
		SELECT id, name, email, class, subject, created_at, updated_at
		FROM teacher WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var teacher models.Teacher
	err = h.db.DB.QueryRow(ctx, selectQuery, id).Scan(
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
			return
		}
		http.Error(w, "Failed to update teacher.", http.StatusInternalServerError)
		log.Printf("Database error during update: %v", err)
		return
	}

	for key, value := range input {
		switch key {
		case "name":
			if name, ok := value.(string); ok {
				teacher.Name = name
			}
		case "email":
			if email, ok := value.(string); ok {
				teacher.Email = email
			}
		case "class":
			if class, ok := value.(string); ok {
				teacher.Class = class
			}
		case "subject":
			if subject, ok := value.(string); ok {
				teacher.Subject = subject
			}
		}
	}

	const updateQuery = `
		UPDATE teacher
		SET name = $1, email = $2, class = $3, subject = $4
		WHERE id = $5
	`

	_, err = h.db.DB.Exec(ctx, updateQuery,
		teacher.Name, teacher.Email, teacher.Class, teacher.Subject, id,
	)
	if err != nil {
		log.Printf("Failed to update teacher with ID %d: %v", id, err)
		http.Error(w, "Failed to update teacher.", http.StatusInternalServerError)
		return
	}

	response := APIResponse[models.Teacher]{Success: true, Data: teacher}
	writeJSON(w, http.StatusOK, response)
}

// ==========================================
// PATCH /teachers/{id}
// ==========================================
func (h *TeacherHandler) PartialUpdateTeacher(w http.ResponseWriter, r *http.Request) {
	var input map[string]any

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}
	defer r.Body.Close()

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Teacher ID must be a valid number.", http.StatusBadRequest)
		return
	}

	const selectQuery = `
		SELECT id, name, email, class, subject, created_at, updated_at
		FROM teacher WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var teacher models.Teacher
	err = h.db.DB.QueryRow(ctx, selectQuery, id).Scan(
		&teacher.ID,
		&teacher.Name,
		&teacher.Email,
		&teacher.Class,
		&teacher.Subject,
		&teacher.CreatedAt,
		&teacher.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Teacher not found.", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch data from the database.", http.StatusInternalServerError)
		return
	}

	teacherValue := reflect.ValueOf(&teacher).Elem()
	teacherType := teacherValue.Type()

	for key, val := range input {
		for i := 0; i < teacherValue.NumField(); i++ {
			field := teacherType.Field(i)
			if field.Tag.Get("json") == key && teacherValue.Field(i).CanSet() {
				teacherValue.Field(i).Set(reflect.ValueOf(val).Convert(teacherValue.Field(i).Type()))
			}
		}
	}

	const updateQuery = `
		UPDATE teacher
		SET name = $1, email = $2, class = $3, subject = $4
		WHERE id = $5
	`

	_, err = h.db.DB.Exec(ctx, updateQuery,
		teacher.Name, teacher.Email, teacher.Class, teacher.Subject, id,
	)
	if err != nil {
		log.Printf("Failed to update teacher with ID %d: %v", id, err)
		http.Error(w, "Failed to update teacher.", http.StatusInternalServerError)
		return
	}

	response := APIResponse[models.Teacher]{Success: true, Data: teacher}
	writeJSON(w, http.StatusOK, response)
}

// ==========================================
// DELETE /teachers/{id}
// ==========================================
func (h *TeacherHandler) DeleteTeacher(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Teacher ID must be a valid number.", http.StatusBadRequest)
		return
	}

	const deleteQuery = `DELETE FROM teacher WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.db.DB.Exec(ctx, deleteQuery, id)
	if err != nil {
		log.Printf("Error deleting teacher with ID %d: %v", id, err)
		http.Error(w, "Internal server error while deleting teacher.", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, fmt.Sprintf("No teacher found with ID %d.", id), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": fmt.Sprintf("Teacher with ID %d successfully deleted.", id),
	})
}

// ==========================================
// PATCH /teachers
// ==========================================
func (h *TeacherHandler) PartialUpdateTeachers(w http.ResponseWriter, r *http.Request) {
	var updates []map[string]any

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}
	defer r.Body.Close()

	ctxR := r.Context()

	conn, err := h.pool.Acquire(ctxR)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Failed to acquire connection: %v", err)
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(ctxR)
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		log.Printf("Failed to begin transaction: %v", err)
		return
	}
	defer func() {
		_ = tx.Rollback(ctxR)
	}()

	txQueries := h.db.WithTx(tx)

	const selectQuery = `
		SELECT id, name, email, class, subject, created_at, updated_at
		FROM teacher
		WHERE id = $1
	`

	const updateQuery = `
		UPDATE teacher
		SET name = $1, email = $2, class = $3, subject = $4
		WHERE id = $5
		RETURNING id, name, email, class, subject, created_at, updated_at
	`

	var updatedTeachers []models.Teacher

	for _, update := range updates {
		idFloat, ok := update["id"].(float64)
		if !ok {
			http.Error(w, "Invalid teacher ID type", http.StatusBadRequest)
			return
		}
		id := int(idFloat)

		ctx, cancel := context.WithTimeout(ctxR, 5*time.Second)

		var teacher models.Teacher
		err = txQueries.DB.QueryRow(ctx, selectQuery, id).Scan(
			&teacher.ID,
			&teacher.Name,
			&teacher.Email,
			&teacher.Class,
			&teacher.Subject,
			&teacher.CreatedAt,
			&teacher.UpdatedAt,
		)
		cancel()
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, fmt.Sprintf("Teacher with ID %d not found", id), http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to fetch teacher", http.StatusInternalServerError)
			log.Printf("QueryRow error: %v", err)
			return
		}

		teacherVal := reflect.ValueOf(&teacher).Elem()
		teacherType := teacherVal.Type()

		for key, val := range update {
			if key == "id" {
				continue
			}

			for i := 0; i < teacherVal.NumField(); i++ {
				field := teacherType.Field(i)
				if field.Tag.Get("json") == key {
					fieldVal := teacherVal.Field(i)
					if fieldVal.CanSet() {
						v := reflect.ValueOf(val)
						if v.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(v.Convert(fieldVal.Type()))
						} else {
							http.Error(w, fmt.Sprintf("Invalid type for field %s", key), http.StatusBadRequest)
							return
						}
					}
					break
				}
			}
		}

		ctx, cancel = context.WithTimeout(ctxR, 5*time.Second)
		defer cancel()

		err = txQueries.DB.QueryRow(ctx, updateQuery,
			teacher.Name, teacher.Email, teacher.Class, teacher.Subject, id,
		).Scan(
			&teacher.ID,
			&teacher.Name,
			&teacher.Email,
			&teacher.Class,
			&teacher.Subject,
			&teacher.CreatedAt,
			&teacher.UpdatedAt,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update teacher with ID %d", id), http.StatusInternalServerError)
			log.Printf("Update failed: %v", err)
			return
		}

		updatedTeachers = append(updatedTeachers, teacher)
	}

	if err := tx.Commit(ctxR); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		log.Printf("Commit error: %v", err)
		return
	}

	response := APIResponse[[]models.Teacher]{Success: true, Data: updatedTeachers}
	writeJSON(w, http.StatusOK, response)
}

// ==========================================
// PATCH /teachers
// ==========================================
func (h *TeacherHandler) DeleteTeachers(w http.ResponseWriter, r *http.Request) {
	var ids []int

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&ids); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}
	defer r.Body.Close()

	ctxR := r.Context()

	conn, err := h.pool.Acquire(ctxR)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Failed to acquire connection: %v", err)
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(ctxR)
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		log.Printf("Failed to begin transaction: %v", err)
		return
	}
	defer func() {
		_ = tx.Rollback(ctxR)
	}()

	txQueries := h.db.WithTx(tx)

	const deleteQuery = `
		DELETE FROM teacher WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var deletedIds []interface{}
	for _, id := range ids {
		result, err := txQueries.DB.Exec(ctx, deleteQuery, id)
		if err != nil {
			log.Printf(":%v", err)
			http.Error(w, "Error deleting teachers.", http.StatusInternalServerError)
			return
		}

		rowsAffected := result.RowsAffected()

		if rowsAffected > 0 {
			deletedIds = append(deletedIds, id)
		}
	}

	if err := tx.Commit(ctxR); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		log.Printf("Commit error: %v", err)
		return
	}

	response := APIResponse[[]any]{Success: true, Data: deletedIds}
	writeJSON(w, http.StatusOK, response)
}

// ==========================================
// Helper Functions
// ==========================================
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

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}
