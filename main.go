package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"pkgmanager/internal/router"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r := router.GetRouter()
	// Start server
	fmt.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":"+port, r))
}
