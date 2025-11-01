package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Teacher struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Class   string `json:"class"`
	Subject string `json:"subject"`
}

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
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
		if err := r.ParseForm(); err != nil {
			fmt.Println("Error parsing form:", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		fmt.Println("Parsed Form Data:")
		for key, values := range r.Form {
			for _, value := range values {
				fmt.Fprintf(w, "%s: %s\n", key, value)
				fmt.Printf("%s: %s\n", key, value)
			}
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Error reading body:", err)
			http.Error(w, "Error reading body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		fmt.Println("Raw Body:", string(body))

		var teacher Teacher
		if err := json.Unmarshal(body, &teacher); err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
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
