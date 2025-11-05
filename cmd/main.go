package main

import (
	"fmt"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World\n")
}

func main() {

	http.HandleFunc("/hello", Handler)

	http.ListenAndServe(":8080", nil)
}
