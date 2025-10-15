package routes

import (
	"net/http"

	"github.com/krsnmvk/gorestapi/internal/api/handlers"
)

func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/teachers", handlers.TeachersHandler)
	mux.HandleFunc("/students", handlers.StudentsHandler)
	mux.HandleFunc("/execs", handlers.ExecsHandler)

	return mux
}
