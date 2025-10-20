package routes

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krsnmvk/gorestapi/internal/api/handlers"
	"github.com/krsnmvk/gorestapi/internal/database"
)

func RegisterRoutes(queries *database.Queries, pool *pgxpool.Pool) *http.ServeMux {
	mux := http.NewServeMux()

	teachers := handlers.NewTeacherHandler(queries, pool)

	mux.HandleFunc("/", handlers.RootHandler)

	mux.HandleFunc("GET /teachers", teachers.GetTeachers)
	mux.HandleFunc("POST /teachers", teachers.CreateTeacher)
	mux.HandleFunc("PATCH /teachers", teachers.PartialUpdateTeachers)
	mux.HandleFunc("GET /teachers/{id}", teachers.GetTeacher)
	mux.HandleFunc("PATCH /teachers/{id}", teachers.PartialUpdateTeacher)
	mux.HandleFunc("PUT /teachers/{id}", teachers.UpdateTeacher)
	mux.HandleFunc("DELETE /teachers/{id}", teachers.DeleteTeacher)

	mux.HandleFunc("/students/", handlers.StudentsHandler)

	mux.HandleFunc("/execs/", handlers.ExecsHandler)

	return mux
}
