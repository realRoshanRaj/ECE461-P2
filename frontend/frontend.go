package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"pkgmanager/internal/models"
	"pkgmanager/pkg/utils"
	"strconv"
)

var baseURL = "https://ece461-project2-2shruw53aq-uc.a.run.app"

// "http://localhost:8080"

func handleError(w http.ResponseWriter, r *http.Request, error_code string, err error) { // Redirects to error page
	fmt.Println(err)

	// Display an error page to the user along with the error code
	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Message string
	}{
		Message: "A " + error_code + " error occurred.",
	}

	// Execute the html
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleSuccess(w http.ResponseWriter, r *http.Request) { // Redirects to success page
	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/success.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func RenderIndex(w http.ResponseWriter, r *http.Request) {
	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func RenderUpdate(w http.ResponseWriter, r *http.Request) {
	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/update.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Get the values of the different fields
	name := r.FormValue("Name")
	version := r.FormValue("Version")
	id := r.FormValue("ID")
	chk := utils.CheckValidChars(id)
	// Make sure the input is valid
	if chk != 1 {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), nil)
	}
	URL := r.FormValue("URL")
	Content := r.FormValue("Content")
	JSProgram := r.FormValue("JSProgram")

	// Make the request body
	bdy_metadata := models.Metadata{Name: name, Version: version, ID: id}
	bdy_data := models.PackageData{URL: URL, Content: Content, JSProgram: JSProgram}
	bdy := models.PackageInfo{Metadata: bdy_metadata, Data: bdy_data}

	// Turn the body into a Json object
	bod, err := json.Marshal(bdy)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Call the update endpoint and pass in the body
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, baseURL+"/package/"+id, bytes.NewBuffer(bod))
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
		return
	}

	// Redirect the user
	handleSuccess(w, r)
}

func RenderCreate(w http.ResponseWriter, r *http.Request) {
	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/create.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Get the values of the input fields
	URL := r.FormValue("URL")
	Content := r.FormValue("Content")
	JSProgram := r.FormValue("JSProgram")
	// Make the request body depending on what fields the user inputs
	bdy := make(map[string]string)
	if URL != "" {
		bdy["URL"] = URL
		bdy["JSProgram"] = JSProgram
	} else {
		bdy["Content"] = Content
		bdy["JSProgram"] = JSProgram
	}

	// Turn the request body into Json
	bod, err := json.Marshal(bdy)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Call the create endpoint and pass in the body
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/package/", bytes.NewBuffer(bod))
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusCreated {
		handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
		return
	}

	// Redirect the user
	handleSuccess(w, r)
}

func RenderRemove(w http.ResponseWriter, r *http.Request) {
	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/remove.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleRemove(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Get the method of removal
	searchBy := r.FormValue("searchBy")
	if searchBy == "id" { // Remove by id
		// Get the id from the user
		id := r.FormValue("value")
		// Check if the id has any invalid characters
		chk := utils.CheckValidChars(id)
		if chk != 1 {
			handleError(w, r, fmt.Sprint(http.StatusBadRequest), nil)
		}

		// Call the delete endpoint with the id
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodDelete, baseURL+"/package/"+id, nil)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}
		defer resp.Body.Close()

		// Check the response status code
		if resp.StatusCode != http.StatusOK {
			handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
			return
		}
	} else { // Remove by name
		// Get the name from the user
		name := r.FormValue("value")
		// Check if the name has any invalid characters
		chk := utils.CheckValidChars(name)
		if chk != 1 {
			handleError(w, r, fmt.Sprint(http.StatusBadRequest), nil)
		}

		// Call the delete by name endpoint and pass in the name
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodDelete, baseURL+"/package/byName/"+name, nil)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}
		defer resp.Body.Close()

		// Check the response status code
		if resp.StatusCode != http.StatusOK {
			handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
			return
		}
	}

	// Redirect the user
	handleSuccess(w, r)
}

func RenderReset(w http.ResponseWriter, r *http.Request) {
	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/reset.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleReset(w http.ResponseWriter, r *http.Request) {
	// Parse the form
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Call the reset endpoint
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, baseURL+"/reset", nil)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
		return
	}

	// Redirect the user to the homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func RenderRate(w http.ResponseWriter, r *http.Request) {
	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/rate.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleRate(w http.ResponseWriter, r *http.Request) {
	// Parse the form
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Get the id from the user
	id := r.FormValue("id")
	// Check if the id has any invalid characters
	chk := utils.CheckValidChars(id)
	if chk != 1 {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), nil)
	}

	// Call the rate endpoint with the package id
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, baseURL+"/package/"+id+"/rate", nil)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
		return
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	defer resp.Body.Close()

	// Convert the Json body into a metrics struct
	var met_data models.Metric
	err = json.Unmarshal(body, &met_data)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Initialize a float64 array that can be passed to the html
	metrics := []float64{
		met_data.NetScore,
		met_data.BusFactor,
		met_data.Correctness,
		met_data.RampUp,
		met_data.ResponsiveMaintainer,
		met_data.LicenseScore,
		met_data.GoodPinningPractice,
		met_data.PullRequest,
	}

	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/rate_results.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, metrics); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func RenderSearch(w http.ResponseWriter, r *http.Request) {
	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/search.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Get the search type and query
	searchType := r.FormValue("type")

	client := &http.Client{}

	// Handle according to search type
	switch searchType {
	case "regex": // Search by regex (no pagination)
		// Get user input and handle in regex function
		query := r.FormValue("regex")
		handleRegex(w, r, client, query)
	case "semver": // Search by queries (with pagination)
		// Get user input and handle in semver function
		offset := r.FormValue("offset")
		name := r.FormValue("name")
		version := r.FormValue("version")
		handleSemver(w, r, client, name, version, offset)
	default: // Handle fall through case
		return
	}
}

func handleRegex(w http.ResponseWriter, r *http.Request, client *http.Client, query string) { // Handling the regex functionality of search
	// Create a request body with the regex
	reqBody, err := json.Marshal(map[string]string{"RegEx": query})
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Make a call to the regex endpoint with the request body
	req, err := http.NewRequest(http.MethodPost, baseURL+"/package/byRegEx", bytes.NewBuffer(reqBody))
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
		return
	}

	// Get the response from the API call
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Store all of the package metadata in a list of maps (maps as we need to add the popularity rating to each package)
	var packages []map[string]string
	if err := json.Unmarshal(respBody, &packages); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Goes through each package and adds the popularity rating
	for _, pkg := range packages {
		// Gets the name to pass to the popularity endpoint
		name := pkg["Name"]

		// Makes a call to the popularity endpoint
		req, err := http.NewRequest(http.MethodGet, baseURL+"/popularity/"+name, nil)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}
		defer resp.Body.Close()

		// Check the status code of the response
		if resp.StatusCode != http.StatusOK {
			handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
			return
		}

		// Read the response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}

		// Get the popularity rating and add it to the package info
		rating := string(body)
		pkg["Rating"] = rating
	}

	// Display an error message to the user if parsing fails
	tmpl, err := template.New("results_regex.html").ParseFiles("templates/results_regex.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	err = tmpl.Execute(w, map[string]interface{}{"Packages": packages})
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func handleSemver(w http.ResponseWriter, r *http.Request, client *http.Client, name string, version string, offset string) { // Handling the semver functionality of search
	// Convert the offset value to an integer
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Create a query with the input from the user and convert it to JSON
	reqQuery := []models.PackageQuery{{Name: name, Version: version}} // Only allows one query at a time for simplicity
	reqBody, err := json.Marshal(reqQuery)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Call the packages endpoint with the request body and appropriate offset
	req, err := http.NewRequest(http.MethodPost, baseURL+"/packages?offset="+offset, bytes.NewBuffer(reqBody))
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	defer resp.Body.Close()

	// Check the status code of the response
	if resp.StatusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
		return
	}

	// Read the response body which has all of the package metadatas
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Store all of the package metadata in a list of maps (maps as we need to add the popularity rating to each package)
	var packages []map[string]string
	err = json.Unmarshal(respBody, &packages)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Goes through each package and adds the popularity rating
	for _, pkg := range packages {
		// Gets the name to pass to the popularity endpoint
		name := pkg["Name"]

		// Calls the popularity endpoint with the name
		req, err := http.NewRequest(http.MethodGet, baseURL+"/popularity/"+name, nil)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}
		defer resp.Body.Close()

		// Check the status code of the response
		if resp.StatusCode != http.StatusOK {
			handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
			return
		}

		// Reads the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}

		// Gets the popularity rating from the body and adds it to the package map
		rating := string(body)
		pkg["Rating"] = rating
	}

	// Display an error message to the user if parsing fails
	tmpl, err := template.New("results_semver.html").ParseFiles("templates/results_semver.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	err = tmpl.Execute(w, map[string]interface{}{"Packages": packages, "Page": offsetInt, "QName": name, "QVersion": version, "Type": "semver", "Sub": func(a int, b int) int { return a - b }, "Add": func(a int, b int) int { return a + b }})
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func RenderHistory(w http.ResponseWriter, r *http.Request) {
	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/history_search.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleHistory(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Get the name
	name := r.FormValue("name")
	// Check that the name is valid
	chk := utils.CheckValidChars(name)
	if chk != 1 {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), nil)
	}

	// Make a call to the get history by name endpoint
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, baseURL+"/package/byName/"+name, nil)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Check the status of the response
	if resp.StatusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Convert the response from a Json object to a list containing the history
	var packageHistories []models.ActionEntry
	err = json.Unmarshal(respBody, &packageHistories)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Iterate through each action and add it to the list of maps that is easier to read in the html
	var packageHistoriesMaps []map[string]interface{}
	for _, ph := range packageHistories {
		phMap := map[string]interface{}{
			"UserName":       ph.User["name"],
			"UserIsAdmin":    ph.User["isAdmin"],
			"Date":           ph.Date,
			"PackageName":    ph.Metadata.Name,
			"PackageVersion": ph.Metadata.Version,
			"PackageID":      ph.Metadata.ID,
			"Action":         ph.Action,
		}
		packageHistoriesMaps = append(packageHistoriesMaps, phMap)
	}

	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/history.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, packageHistoriesMaps); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func RenderDownload(w http.ResponseWriter, r *http.Request) {
	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/download_search.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Get the id
	id := r.FormValue("id")
	// Check if the id is valid
	chk := utils.CheckValidChars(id)
	if chk != 1 {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), nil)
	}

	// Make a request to the package endpoint with the package id
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, baseURL+"/package/"+id, nil)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Check the response code
	if resp.StatusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
		return
	}
	defer resp.Body.Close()

	// Stream the response body and decode the JSON data in chunks to handle the possibly large content
	var packageInfo map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&packageInfo); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Get the zip file
	zip := packageInfo["data"].(map[string]interface{})["Content"].(string)
	// Find it's size
	zipSize, statusCode := utils.GetZipSize(zip)
	if statusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(statusCode), err)
		return
	}
	// Add it to the map to be displayed to the user
	packageInfo["Size"] = zipSize

	// Get the name of the package
	name := packageInfo["metadata"].(map[string]interface{})["Name"].(string)

	// Use the name to get the package popularity by call the popularity endpoint
	req, err = http.NewRequest(http.MethodGet, baseURL+"/popularity/"+name, nil)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	resp, err = client.Do(req)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Check the response code
	if resp.StatusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
		return
	}
	defer resp.Body.Close()

	// Read the response into a string
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	rating := string(body)

	// Add the popularity to the map to display to the user
	packageInfo["Rating"] = rating

	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/download.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, packageInfo); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func RenderCreateReview(w http.ResponseWriter, r *http.Request) {
	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/create_review.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleCreateReview(w http.ResponseWriter, r *http.Request) {
	//Parse the form
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Get the values of the input fields
	userName := r.FormValue("userName")
	packageName := r.FormValue("packageName")
	stars := r.FormValue("stars")
	review := r.FormValue("review")

	// Create the request body
	bdy := map[string]string{"userName": userName, "packageName": packageName, "stars": stars, "review": review}
	bod, err := json.Marshal(bdy)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Call the create review endpoint with the request body
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/package/review", bytes.NewBuffer(bod))
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusCreated {
		handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
		return
	}

	// Redirect the user
	handleSuccess(w, r)
}

func RenderDeleteReview(w http.ResponseWriter, r *http.Request) {
	// Display an error message to the user if parsing fails
	tmpl, err := template.ParseFiles("templates/delete_review.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	// Execute the html
	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleDeleteReview(w http.ResponseWriter, r *http.Request) {
	//Parse the form
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Get the values of the input fields
	userName := r.FormValue("userName")
	packageName := r.FormValue("packageName")

	// Create the request body as a Json object
	bdy := map[string]string{"userName": userName, "packageName": packageName}
	bod, err := json.Marshal(bdy)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Call the delete review endpoint with the request body
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, baseURL+"/package/review", bytes.NewBuffer(bod))
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(resp.StatusCode), nil)
		return
	}

	// Redirect the user
	handleSuccess(w, r)
}
