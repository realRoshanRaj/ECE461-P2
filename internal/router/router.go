package router

import (
	"net/http"

	handler "pkgmanager/internal/handlers"

	"pkgmanager/frontend"

	"github.com/apsystole/log"
	"github.com/go-chi/chi"
)

func GetRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/", frontend.RenderIndex)
	router.Get("/update", frontend.RenderUpdate)
	router.Post("/update", frontend.HandleUpdate)
	router.Get("/create", frontend.RenderCreate)
	router.Post("/create", frontend.HandleCreate)
	router.Get("/remove", frontend.RenderRemove)
	router.Post("/remove", frontend.HandleRemove)
	router.Get("/reset", frontend.RenderReset)
	router.Post("/reset", frontend.HandleReset)

	// Define endpointss
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
	log.Println("Route not implemented yet", r.Body)
	w.WriteHeader(http.StatusNotImplemented)
}
