package handlers

import (
	"fmt"
	"net/http"
)

func StudentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Students Route"))
	fmt.Println("Hello Students Route")
}
