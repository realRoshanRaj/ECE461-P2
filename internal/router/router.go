package router

import (
	"net/http"

	"pkgmanager/frontend"
	handler "pkgmanager/internal/handlers"

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
	router.Get("/rate", frontend.RenderRate)
	router.Post("/rate", frontend.HandleRate)
	router.Get("/search", frontend.RenderSearch)
	router.Post("/search", frontend.HandleSearch)
	router.Get("/history", frontend.RenderHistory)
	router.Post("/history", frontend.HandleHistory)
	router.Get("/download", frontend.RenderDownload)
	router.Post("/download", frontend.HandleDownload)
	router.Get("/create_review", frontend.RenderCreateReview)
	router.Post("/create_review", frontend.HandleCreateReview)
	router.Get("/delete_review", frontend.RenderDeleteReview)
	router.Post("/delete_review", frontend.HandleDeleteReview)

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

	router.Get("/popularity/{name}", handler.GetPackagePopularity)
	router.Post("/package/review", handler.ReviewPackage)
	router.Delete("/package/review", handler.DeleteReview)

	router.Post("/package/byRegEx", handler.GetPackageByRegex)

	return router
}

func tempRoute(w http.ResponseWriter, r *http.Request) {
	log.Println("Route not implemented yet")
	w.WriteHeader(http.StatusNotImplemented)
}
