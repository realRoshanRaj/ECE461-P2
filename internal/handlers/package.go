package handlers

import (
	"encoding/json"
	"net/http"
	"pkgmanager/internal/metrics"
	"pkgmanager/internal/models"
	"pkgmanager/pkg/db"
	"pkgmanager/pkg/utils"

	"github.com/go-chi/chi"
)

func CreatePackage(w http.ResponseWriter, r *http.Request) {
	// initialize a packagedata struct based on the request body
	packageData := models.PackageData{}
	err := json.NewDecoder(r.Body).Decode(&packageData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// TODO: find actual metadata
	// metadata := models.Metadata{Name: "package_Name", Version: "package_Version", ID: "packageData_ID"}
	var metadata models.Metadata
	if packageData.Content == "" && packageData.URL != "" {
		// URL method
		metadata = utils.ExtractMetadataFromURL(packageData.URL)
	} else if packageData.Content != "" && packageData.URL == "" {
		// Content method (zip file)
		var foundPackageJson bool
		metadata, foundPackageJson = utils.ExtractMetadataFromZip(packageData.Content)
		if !foundPackageJson {
			w.WriteHeader(http.StatusBadRequest) // 400
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// Initialize package info struct that uses packagedata struct
	packageInfo := models.PackageInfo{
		Data:     packageData,
		Metadata: metadata,
	}
	// TODO: http.StatusFailedDependency (424) if package rating doesn't meet requirements
	rating := metrics.GenerateMetrics(packageInfo.Metadata.Repository)
	if !metrics.IsRatingQualified(rating) {
		w.WriteHeader(http.StatusFailedDependency) // 424
		return
	}
	// Create package in database
	_, statusCode := db.CreatePackage(&packageInfo)

	if statusCode == http.StatusCreated {
		responseJSON(w, http.StatusCreated, packageInfo)
	} else {
		w.WriteHeader(statusCode) // handles the 409 conflict error
	}
}

func DownloadPackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "id")

	pkgInfo, statusCode := db.GetPackageByID(packageID)
	if statusCode == http.StatusOK {
		responseJSON(w, http.StatusOK, pkgInfo)
	} else {
		w.WriteHeader(statusCode) // handles the 404 error
	}
}

func UpdatePackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "id")
	// initialize a packagedata struct based on the request body
	packageInfo := models.PackageInfo{}
	err := json.NewDecoder(r.Body).Decode(&packageInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// TODO: find actual metadata

	statusCode := db.UpdatePackageByID(packageID, packageInfo)
	w.WriteHeader(statusCode)
	// responseJSON(w, http.StatusCreated, packageID)
}

func DeletePackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "id")
	statusCode := db.DeletePackageByID(packageID)
	w.WriteHeader(statusCode) // handles error/status codes
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

func GetPackageHistoryByName(w http.ResponseWriter, r *http.Request) {
	packageName := chi.URLParam(r, "name")
	pkgHistory, statusCode := db.GetPackageHistoryByName(packageName)
	if statusCode == http.StatusOK {
		responseJSON(w, http.StatusOK, pkgHistory)
	} else {
		w.WriteHeader(statusCode) // handles the 404 error
	}
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
