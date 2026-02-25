package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := "8081"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	// Serve static files from the demo directory
	fs := http.FileServer(http.Dir("demo"))
	http.Handle("/", fs)

	fmt.Printf("🚀 Demo Client UI started at http://localhost:%s\n", port)
	fmt.Printf("👉 Open http://localhost:%s in your browser\n", port)
	fmt.Println("⚠️  Make sure the backend API is running at http://localhost:8080")

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
