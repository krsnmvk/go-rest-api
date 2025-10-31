package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := ":8080"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello Root Route"))
		fmt.Println("Hello Root Route")
	})

	http.HandleFunc("/teachers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte("Hello GET Method on Teachers Route"))
			fmt.Println("Hello GET Method on Teachers Route")

		case http.MethodPost:
			w.Write([]byte("Hello POST Method on Teachers Route"))
			fmt.Println("Hello POST Method on Teachers Route")

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

	})

	http.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello Students Route"))
		fmt.Println("Hello Students Route")
	})

	http.HandleFunc("/execs", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello Execs Route"))
		fmt.Println("Hello Execs Route")
	})

	log.Println("Server runing on http://localhost:8080")
	if err := http.ListenAndServe(port, nil); err != nil {
		panic(fmt.Sprintf("Error starting the server: %v\n", err))
	}
}
