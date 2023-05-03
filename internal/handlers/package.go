package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"pkgmanager/internal/metrics"
	"pkgmanager/internal/models"
	"pkgmanager/pkg/db"
	"pkgmanager/pkg/utils"
	"strconv"

	"github.com/apsystole/log"
	"github.com/go-chi/chi"
)

func CreatePackage(w http.ResponseWriter, r *http.Request) {
	// Initialize a packagedata struct regardless of whether Content or URL being used
	packageData := models.PackageData{}

	// Load the request body into a package data struct
	body, err := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	err = json.NewDecoder(r.Body).Decode(&packageData)

	// Debug statements telling us about the input package
	log.Debugln(string(body))
	log.Debugf("CreatePackage called %+v", packageData)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	var metadata models.Metadata
	var contentTooBig bool = false
	// Checking what the type of the input is, URL or Content
	if packageData.Content == "" && packageData.URL != "" {
		// URL method first
		rating := metrics.GenerateMetrics(packageData.URL)
		// Checking if it scores high enough on our metrics
		log.Printf("Package Ingestion Rating: %+v\n", rating)
		if !metrics.MeasureIngestibility(rating) {
			w.WriteHeader(http.StatusFailedDependency) // 424
			return
		}
		// Next we get the metadata from the URL
		var found bool
		metadata, found = utils.ExtractMetadataFromURL(packageData.URL)
		if !found {
			w.WriteHeader(http.StatusBadRequest) // 400
			return
		}
	} else if packageData.Content != "" && packageData.URL == "" {
		// Content method (zip file)
		var foundPackageJson bool
		// Here we extract the metadata from the zip file
		metadata, foundPackageJson, contentTooBig = utils.ExtractMetadataFromZip(packageData.Content)
		if !foundPackageJson {
			w.WriteHeader(http.StatusBadRequest) // 400
			return
		}
	} else {
		// If both the zip file and url provided
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// Initialize package info struct that uses packagedata and metadata struct
	packageInfo := models.PackageInfo{
		Data:     packageData,
		Metadata: metadata,
	}
	// Log the metadata
	log.Printf("Create: %+v\n", packageInfo.Metadata)

	// Create package in database
	_, statusCode := db.CreatePackage(&packageInfo, contentTooBig)

	// Respond appropriately
	if statusCode == http.StatusCreated {
		responseJSON(w, http.StatusCreated, packageInfo)
	} else {
		w.WriteHeader(statusCode) // handles the 409 conflict error
	}
}

func DownloadPackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "id")
	log.Debugf("DownloadPackage called %s", packageID)
	// Gets the package from the database
	pkgInfo, statusCode := db.GetPackageByID(packageID, 1)
	if statusCode == http.StatusOK {
		// If there is no content in the package, then download the content from the URL
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
	// Initialize a packagedata struct based on the request body
	packageInfo := models.PackageInfo{}
	err := json.NewDecoder(r.Body).Decode(&packageInfo)
	log.Debugf("UpdatePackage (%s) called %+v", packageID, packageInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// Find and set the repository for the new package
	if packageInfo.Data.Content == "" && packageInfo.Data.URL != "" {
		// given update url
		packageInfo.Metadata.Repository = packageInfo.Data.URL
	} else if packageInfo.Data.Content != "" && packageInfo.Data.URL == "" {
		// Content method (zip file)
		// Find the repository for the new package
		metadata, foundPackageJson, _ := utils.ExtractMetadataFromZip(packageInfo.Data.Content)
		if !foundPackageJson {
			w.WriteHeader(http.StatusBadRequest) // 400
			return
		}
		packageInfo.Metadata.Repository = metadata.Repository
	} else {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	// Update the package in the database
	statusCode := db.UpdatePackageByID(packageID, packageInfo)
	w.WriteHeader(statusCode)
}

func DeletePackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "id")
	// Delete the package in the database
	statusCode := db.DeletePackageByID(packageID)
	w.WriteHeader(statusCode) // handles error/status codes
}

func RatePackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "id")
	log.Debugf("RatePackage called %s", packageID)
	// Getting the package from the database
	pkgInfo, statusCode := db.GetPackageByID(packageID, 0)
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode) // handles the 404 error
		return
	}

	// Sending the URL to get the metrics
	metrics := metrics.GenerateMetrics(pkgInfo.Metadata.Repository)
	responseJSON(w, http.StatusOK, metrics)
}

func GetPackageHistoryByName(w http.ResponseWriter, r *http.Request) {
	packageName := chi.URLParam(r, "name")
	// Get all the package history from the database with the given name
	pkgHistory, statusCode := db.GetPackageHistoryByName(packageName)
	if statusCode == http.StatusOK {
		responseJSON(w, http.StatusOK, pkgHistory)
	} else {
		w.WriteHeader(statusCode) // handles the 404 error
	}
}

func DeletePackageByName(w http.ResponseWriter, r *http.Request) {
	packageName := chi.URLParam(r, "name")
	// Deletes all packages with this name
	statusCode := db.DeletePackageByName(packageName)
	w.WriteHeader(statusCode) // handles error/status codes
}

func CreateReview(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}

	// Convert the request body from a json to a map
	var requestBod map[string]string
	err = json.Unmarshal(reqBody, &requestBod)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	// Get the review elements from the request
	userName := requestBod["userName"]
	stars, err := strconv.Atoi(requestBod["stars"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	if stars < 0 || stars > 5 {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	review := requestBod["review"]
	packageName := requestBod["packageName"]

	// Create the review in the database
	statusCode := db.CreateReview(userName, stars, review, packageName)

	if statusCode == http.StatusCreated {
		responseJSON(w, http.StatusCreated, requestBod)
	} else {
		w.WriteHeader(statusCode)
	}
}

func DeleteReview(w http.ResponseWriter, r *http.Request) {
	// Get the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}

	// Convert the Json into a map
	var requestBod map[string]string
	err = json.Unmarshal(reqBody, &requestBod)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	// Get the review identifiers
	userName := requestBod["userName"]
	packageName := requestBod["packageName"]

	//Delete the review from the database
	statusCode := db.DeleteReview(userName, packageName)

	if statusCode == http.StatusOK {
		responseJSON(w, http.StatusOK, requestBod)
	} else {
		w.WriteHeader(statusCode)
	}
}

func GetPackagePopularity(w http.ResponseWriter, r *http.Request) {
	// Get the popularity score by combining the review stars, the github stars, and the number of downloads
	packageName := chi.URLParam(r, "name")
	popularity, statusCode := db.GetPackagePopularityByName(packageName)

	if statusCode == http.StatusOK {
		responseJSON(w, http.StatusOK, popularity)
	} else {
		w.WriteHeader(statusCode) // handles the 404 error
	}
}

func GetPackageByRegex(w http.ResponseWriter, r *http.Request) {
	// Get the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 400
		return
	}

	// Get the regex from the request
	var regexMap map[string]string
	err = json.Unmarshal(reqBody, &regexMap)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	regex := regexMap["RegEx"]

	// Search the database for packages names/readmes matching the regex
	packages, statusCode := db.GetPackageByRegex(string(regex))

	if statusCode == http.StatusOK {
		responseJSON(w, http.StatusOK, packages)
	} else {
		w.WriteHeader(statusCode)
	}
}

func GetPackagesFromQueries(w http.ResponseWriter, r *http.Request) {
	// Page length set to 10
	const MAX_PER_PAGE = 10

	// Get the queries from the request
	var queries []models.PackageQuery
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	err = json.NewDecoder(r.Body).Decode(&queries)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// Get offset from the query parameter
	offsetStr := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 1
	}

	// Get matching packages from the database
	packages, statusCode := db.GetPackagesFromQueries(queries, (offset-1)*MAX_PER_PAGE, MAX_PER_PAGE) // Database takes offset in terms of index not pages
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
		return
	}

	// If no packages matching query return empty response with same offset
	if len(packages) == 0 {
		packages = []models.Metadata{}
		offset--
	}

	w.Header().Set("offset", strconv.Itoa(offset+1))

	responseJSON(w, http.StatusOK, packages)
}

func responseJSON(w http.ResponseWriter, status int, payload interface{}) { // respondJSON makes the response with payload as json format
	// Convert payload to Json object
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Print the payload to log
	prettyPrint, err := json.MarshalIndent(payload, "", "  ")
	log.Debugln(string(prettyPrint))

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

func ResetRegistry(w http.ResponseWriter, r *http.Request) {
	// Delete the packages from the database
	err := db.DeletePackages()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Delete the history from the database
	err = db.DeleteHistory()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Delete the reviews from the database
	err = db.DeleteReviews()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Delete the files that are too large from the database
	err = db.ClearZipStorage()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

}
