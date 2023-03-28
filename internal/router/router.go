package router

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func GetRouter() *chi.Mux {
	r := chi.NewRouter()

	// Define endpoints
	r.Post("/packages", tempRoute)

	r.Delete("/reset", tempRoute)

	r.Post("/package", tempRoute)
	r.Get("/package/{id}", tempRoute)
	r.Put("/package/{id}", tempRoute)
	r.Delete("/package/{id}", tempRoute)
	r.Get("/package/{id}/rate", tempRoute)

	r.Put("/authenticate", tempRoute)

	r.Get("/package/byName/{name}", tempRoute)
	r.Delete("/package/byName/{name}", tempRoute)

	r.Post("/package/byRegEx/{regex}", tempRoute)

	return r
}

func tempRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Route not implemented yet", w.Header())
}
