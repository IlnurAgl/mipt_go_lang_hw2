package main

import (
	"fmt"
	"io"
	"net/http"
)

func ping(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /ping request\n")
	_, err := io.WriteString(w, "pong")
	if err != nil {
		return
	}
}

func main() {
	http.HandleFunc("/ping", ping)

	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
