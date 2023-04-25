package router

import (
	"net/http"

	handler "pkgmanager/internal/handlers"

	"github.com/apsystole/log"
	"github.com/go-chi/chi"
)

func GetRouter() *chi.Mux {
	router := chi.NewRouter()

	// Define endpoints
	router.Post("/packages", handler.GetPackages)

	router.Delete("/reset", handler.ResetRegistry)

	router.Route("/package", func(r chi.Router) {
		r.Post("/", handler.CreatePackage)
		r.Get("/{id}", handler.DownloadPackage)
		r.Put("/{id}", handler.UpdatePackage)
		r.Delete("/{id}", handler.DeletePackage)
		r.Get("/{id}/rate", handler.RatePackage)
	})

	router.Put("/authenticate", tempRoute)

	router.Get("/package/byName/{name}", handler.GetPackageHistoryByName)
	router.Delete("/package/byName/{name}", handler.DeletePackageByName)

	router.Post("/package/byRegEx", handler.GetPackageByRegex)

	return router
}

func tempRoute(w http.ResponseWriter, r *http.Request) {
	log.Println("Route not implemented yet")
	w.WriteHeader(http.StatusNotImplemented)
}
