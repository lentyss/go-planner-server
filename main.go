package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "", 0)

	port := os.Getenv("TODO_PORT")

	if port == "" {
		port = "7540"
	}

	webDir := "./web"
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		logger.Fatalf("Web directory '%s' not exist", webDir)
	}

	http.Handle("/", http.FileServer(http.Dir(webDir)))

	logger.Printf("Starting server on the port %s...\n", port)
	logger.Fatal(http.ListenAndServe(":"+port, nil))
}
