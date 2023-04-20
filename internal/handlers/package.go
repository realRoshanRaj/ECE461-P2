package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pkgmanager/internal/metrics"
	"pkgmanager/internal/models"
	"pkgmanager/pkg/db"
	"pkgmanager/pkg/utils"
	"strconv"

	"strings"

	"github.com/apsystole/log"
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

	// metadata := models.Metadata{Name: "package_Name", Version: "package_Version", ID: "packageData_ID"}
	var metadata models.Metadata
	var contentTooBig bool = false
	if packageData.Content == "" && packageData.URL != "" {
		// URL method
		// TODO: http.StatusFailedDependency (424) if package rating doesn't meet requirements (BUT IS ALWAYS TRUE)
		rating := metrics.GenerateMetrics(packageData.URL)
		log.Printf("%+v\n", rating)
		if !metrics.MeasureIngestibility(rating) {
			w.WriteHeader(http.StatusFailedDependency) // 424
			return
		}
		var found bool
		metadata, found = utils.ExtractMetadataFromURL(packageData.URL)
		if !found {
			w.WriteHeader(http.StatusBadRequest) // 400
			return
		}
		// packageType = 0 // URL method
		// packageData.Content = utils.ExtractZipFromURL(packageData.URL)
	} else if packageData.Content != "" && packageData.URL == "" {
		// Content method (zip file)
		var foundPackageJson bool
		metadata, foundPackageJson, contentTooBig = utils.ExtractMetadataFromZip(packageData.Content)
		// log.Info("length of content: ", strconv.Itoa(len(packageData.Content)))
		if !foundPackageJson {
			w.WriteHeader(http.StatusBadRequest) // 400
			return
		}
		// packageType = 1 // Content method
	} else {
		// Both zip file and url provided
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// Initialize package info struct that uses packagedata struct
	packageInfo := models.PackageInfo{
		Data:     packageData,
		Metadata: metadata,
	}
	//output the packageInfo json to console
	fmt.Printf("%+v\n", packageInfo.Metadata)
	// log.Info("I'm here", unsafe.Sizeof(metadata), unsafe.Sizeof(packageInfo))

	// Create package in database
	_, statusCode := db.CreatePackage(&packageInfo, contentTooBig)

	if statusCode == http.StatusCreated {
		responseJSON(w, http.StatusCreated, packageInfo)
	} else {
		w.WriteHeader(statusCode) // handles the 409 conflict error
	}
}

func DownloadPackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "id")
	// TODO: also need to return the content if URL only exists
	pkgInfo, statusCode := db.GetPackageByID(packageID, 1)
	if statusCode == http.StatusOK {
		// if there is no content in the database, then download the content from the URL
		if pkgInfo.Data.Content == "" {
			pkgInfo.Data.Content = utils.ExtractZipFromURL(pkgInfo.Metadata.Repository)
		}
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

	pkgInfo, statusCode := db.GetPackageByID(packageID, 0)
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode) // handles the 404 error
		return
	}

	metrics := metrics.GenerateMetrics(pkgInfo.Metadata.Repository)
	// if metrics != nil {
	responseJSON(w, http.StatusOK, metrics)
	// } else {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// }

	// payload := []byte(packageID)
	// w.WriteHeader(http.StatusCreated)
	// _, err := w.Write(payload) // put json here
	// if err != nil {
	// 	log.Println(err)
	// }

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

func DeletePackageByName(w http.ResponseWriter, r *http.Request) {
	packageName := chi.URLParam(r, "name")
	statusCode := db.DeletePackageByName(packageName)
	w.WriteHeader(statusCode) // handles error/status codes
}

func GetPackageByRegex(w http.ResponseWriter, r *http.Request) {
	regex, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	packages, statusCode := db.GetPackageByRegex(string(regex))

	if statusCode == http.StatusOK {
		responseJSON(w, http.StatusOK, packages)
	} else {
		w.WriteHeader(statusCode)
	}
}

const MAX_PER_PAGE = 8

func GetPackages(w http.ResponseWriter, r *http.Request) {
	var pkgs []models.PackageQuery
	err := json.NewDecoder(r.Body).Decode(&pkgs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	var Version string
	var name string
	mode := "Exact"
	for _, pkg := range pkgs {
		Version = pkg.Version
		name = pkg.Name
	}

	if strings.Contains(Version, "-") {
		mode = "Bounded range"
	} else if strings.Contains(Version, "^") {
		mode = "Carat"
	} else if strings.Contains(Version, "~") {
		mode = "Tilde"
	}

	pageNumStr := r.URL.Query().Get("query")
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil {
		pageNum = 1
	}

	startIndex := (pageNum - 1) * MAX_PER_PAGE
	endIndex := startIndex + MAX_PER_PAGE

	packages, statusCode := db.GetPackages(Version, name, mode)

	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
		return
	}

	if pageNum <= 0 || startIndex >= len(packages) {
		responseJSON(w, http.StatusOK, []models.PackageQuery{})
		return
	}

	if endIndex > len(packages) {
		endIndex = len(packages)
	}

	responseJSON(w, http.StatusOK, packages[startIndex:endIndex])
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

func ResetRegistry(w http.ResponseWriter, r *http.Request) {

	err := db.DeletePackages()

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = db.DeleteHistory()

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = db.ClearZipStorage()

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

}
