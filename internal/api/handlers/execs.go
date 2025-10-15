package handlers

import (
	"fmt"
	"net/http"
)

func ExecsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Execs Route"))
	fmt.Println("Hello Execs Route")
}
