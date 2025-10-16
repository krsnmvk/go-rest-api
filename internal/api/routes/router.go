package routes

import (
	"net/http"

	"github.com/krsnmvk/gorestapi/internal/api/handlers"
	"github.com/krsnmvk/gorestapi/internal/database"
)

func RegisterRoutes(queries *database.Queries) *http.ServeMux {
	mux := http.NewServeMux()

	teachers := handlers.NewTeacher(queries)

	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/teachers", teachers.TeachersHandler)
	mux.HandleFunc("/students", handlers.StudentsHandler)
	mux.HandleFunc("/execs", handlers.ExecsHandler)

	return mux
}
