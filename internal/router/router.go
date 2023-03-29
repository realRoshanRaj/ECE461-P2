package router

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func GetRouter() *chi.Mux {
	router := chi.NewRouter()

	// Define endpoints
	router.Post("/packages", tempRoute)

	router.Delete("/reset", tempRoute)

	// r.Post("/package", tempRoute)
	// r.Get("/package/{id}", getPackage)
	// r.Put("/package/{id}", tempRoute)
	// r.Delete("/package/{id}", tempRoute)
	// r.Get("/package/{id}/rate", tempRoute)
	router.Route("/package", func(r chi.Router) {
		r.Post("/", postPackage)
		r.Get("/{id}", getPackage)
		r.Put("/{id}", tempRoute)
		r.Delete("/{id}", tempRoute)
		r.Get("/{id}/rate", tempRoute)
	})

	router.Put("/authenticate", tempRoute)

	router.Get("/package/byName/{name}", tempRoute)
	router.Delete("/package/byName/{name}", tempRoute)

	router.Post("/package/byRegEx/{regex}", tempRoute)

	return router
}

func tempRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Route not implemented yet", r.Body)
}

func postPackage(w http.ResponseWriter, r *http.Request) {
}

func getPackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "id")
	payload := []byte(packageID)
	w.WriteHeader(http.StatusCreated)
	_, err := w.Write(payload) // put json here
	if err != nil {
		log.Println(err)
	}
}
