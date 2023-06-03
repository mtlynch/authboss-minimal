package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

func main() {
	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5050"
	}
	log.Printf("Listening on %s", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
