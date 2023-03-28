package main

import (
	"fmt"
	"log"
	"net/http"
	"pkgmanager/internal/router"
)

func main() {
	r := router.GetRouter()
	// Start server
	fmt.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
