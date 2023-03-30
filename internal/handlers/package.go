package handlers

import (
	"encoding/json"
	"net/http"
	"pkgmanager/internal/models"
	"pkgmanager/pkg/db"

	"github.com/go-chi/chi"
)

func CreatePackage(w http.ResponseWriter, r *http.Request) {
	// initialize a packagedata struct based on the request body
	packageData := models.PackageData{}
	err := json.NewDecoder(r.Body).Decode(&packageData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
	}

	// TODO: find actual metadata
	metadata := models.Metadata{Name: "package_Name", Version: "package_Version", ID: "packageData_ID"}

	// Initialize package info struct that uses packagedata struct
	packageInfo := models.PackageInfo{
		Data:     packageData,
		Metadata: metadata,
	}

	// Create package in database
	db.CreatePackageDB(packageInfo)

	// w.Write([]byte("Package created"))
	responseJSON(w, http.StatusCreated, packageInfo)
}

func DownloadPackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "id")
	// payload := []byte(packageID)
	// w.WriteHeader(http.StatusCreated)
	// _, err := w.Write(payload) // put json here
	// if err != nil {
	// 	log.Println(err)
	// }
	responseJSON(w, http.StatusCreated, packageID)
}

func UpdatePackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "id")
	// payload := []byte(packageID)
	// w.WriteHeader(http.StatusCreated)
	// _, err := w.Write(payload) // put json here
	// if err != nil {
	// 	log.Println(err)
	// }
	responseJSON(w, http.StatusCreated, packageID)
}

func DeletePackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "id")
	// payload := []byte(packageID)
	// w.WriteHeader(http.StatusCreated)
	// _, err := w.Write(payload) // put json here
	// if err != nil {
	// 	log.Println(err)
	// }
	responseJSON(w, http.StatusCreated, packageID)
}

func RatePackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "id")
	// payload := []byte(packageID)
	// w.WriteHeader(http.StatusCreated)
	// _, err := w.Write(payload) // put json here
	// if err != nil {
	// 	log.Println(err)
	// }
	responseJSON(w, http.StatusCreated, packageID)
}

// respondJSON makes the response with payload as json format
func responseJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}
