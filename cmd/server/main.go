package main

import (
	"fmt"
	"net/http"
	"pkgmanager/internal/router"
)

func main() {
	r := router.GetRouter()
	// Start server
	fmt.Println("Server started on port 8080")
	http.ListenAndServe(":8080", r)
}
