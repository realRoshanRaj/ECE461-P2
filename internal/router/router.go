package router

import (
	"fmt"
	"net/http"

	handler "pkgmanager/internal/handlers"

	"github.com/go-chi/chi"
)

func GetRouter() *chi.Mux {
	router := chi.NewRouter()

	// Define endpoints
	router.Post("/packages", handler.GetPackages)

	router.Delete("/reset", tempRoute)

	router.Route("/package", func(r chi.Router) {
		r.Post("/", handler.CreatePackage)
		r.Get("/{id}", handler.DownloadPackage)
		r.Put("/{id}", handler.UpdatePackage)
		r.Delete("/{id}", handler.DeletePackage)
		r.Get("/{id}/rate", handler.RatePackage)
	})

	router.Put("/authenticate", tempRoute)

	router.Get("/package/byName/{name}", handler.GetPackageHistoryByName)
	router.Delete("/package/byName/{name}", tempRoute)

	router.Post("/package/byRegEx", handler.GetPackageByRegex)

	return router
}

func tempRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Route not implemented yet", r.Body)
	w.WriteHeader(http.StatusNotImplemented)
}
