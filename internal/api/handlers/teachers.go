package handlers

import (
	"fmt"
	"net/http"
)

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello GET Method on Teachers Route"))
		fmt.Println("Hello GET Method on Teachers Route")

	case http.MethodPost:
		w.Write([]byte("Hello POST Method on Teachers Route"))
		fmt.Println("Hello POST Method on Teachers Route")

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
