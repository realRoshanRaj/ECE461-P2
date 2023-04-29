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
)

// define a struct to hold the data for the template
type PageData struct {
	Title string
	Body  string
}

var baseURL = "http://localhost:8080"

// "https://ece461-project2-2shruw53aq-uc.a.run.app"

// Redirects to error page
func handleError(w http.ResponseWriter, r *http.Request, error_code string, err error) {
	fmt.Println(err)
	// Display an error message to the user
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
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Redirects to success page
func handleSuccess(w http.ResponseWriter, r *http.Request) {
	// Display an error message to the user
	tmpl, err := template.ParseFiles("templates/success.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

// Render function to render the template with the data
func RenderIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	data := PageData{
		Title: "Module Registry",
		Body:  "Welcome to our Module Registry",
	}

	if err := tmpl.Execute(w, data); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func RenderUpdate(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/update.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

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

	// Get the value of the "name" field
	name := r.FormValue("Name")
	version := r.FormValue("Version")
	id := r.FormValue("ID")
	URL := r.FormValue("URL")
	Content := r.FormValue("Content")
	JSProgram := r.FormValue("JSProgram")

	bdy := make(map[string]map[string]string)

	bdy_metadata := map[string]string{"Name": name, "Version": version, "ID": id}
	bdy_data := map[string]string{"URL": URL, "Content": Content, "JSProgram": JSProgram}
	bdy["metadata"] = bdy_metadata
	bdy["data"] = bdy_data

	bod, err := json.Marshal(bdy)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Call the API endpoint
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
		handleError(w, r, fmt.Sprint(resp.StatusCode), err)
		return
	}

	// Redirect the user
	handleSuccess(w, r)
}

func RenderCreate(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/create.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

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

	bdy := make(map[string]string)
	if URL != "" {
		bdy["URL"] = URL
		bdy["JSProgram"] = JSProgram
	} else {
		bdy["Content"] = Content
		bdy["JSProgram"] = JSProgram
	}

	bod, err := json.Marshal(bdy)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Call the API endpoint
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
		handleError(w, r, fmt.Sprint(resp.StatusCode), err)
		return
	}

	// Redirect the user
	handleSuccess(w, r)
}

func RenderRemove(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/remove.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleRemove(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	searchBy := r.FormValue("searchBy")
	if searchBy == "id" {
		id := r.FormValue("value")
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
			handleError(w, r, fmt.Sprint(resp.StatusCode), err)
			return
		}
	} else {
		name := r.FormValue("value")
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
			handleError(w, r, fmt.Sprint(resp.StatusCode), err)
			return
		}
	}

	// Redirect the user
	handleSuccess(w, r)
}

func RenderReset(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/reset.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleReset(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}
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
		handleError(w, r, fmt.Sprint(resp.StatusCode), err)
		return
	}

	// Not Redirecting but Handled
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func RenderRate(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/rate.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func HandleRate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	id := r.FormValue("id")
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
		handleError(w, r, fmt.Sprint(resp.StatusCode), err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	var met_data models.Metric

	err = json.Unmarshal(body, &met_data)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	defer resp.Body.Close()
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

	tmpl, err := template.ParseFiles("templates/rate_results.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	if err := tmpl.Execute(w, metrics); err != nil {
		fmt.Println(err)
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func RenderSearch(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/search.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

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
	query := r.FormValue("q")

	// Create a new HTTP client
	client := &http.Client{}

	// Declare a variable to hold the response body
	var respBody []byte

	// Handle the different search types
	switch searchType {
	case "regex":
		// Create a JSON body for the request
		reqBody, err := json.Marshal(map[string]string{"RegEx": string(query)})
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
			return
		}

		// Make a POST request to the API
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
		if resp.StatusCode != http.StatusOK {
			handleError(w, r, fmt.Sprint(resp.StatusCode), err)
			return
		}
		defer resp.Body.Close()

		// Read the response body
		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}

	case "semver":
		// Get the name and version from the form data
		name := r.FormValue("name")
		version := r.FormValue("version")
		reqQuery := []map[string]string{{"Name": name, "Version": version}}

		// Create a JSON body for the request
		reqBody, err := json.Marshal(reqQuery)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
			return
		}

		req, err := http.NewRequest(http.MethodPost, baseURL+"/packages", bytes.NewBuffer(reqBody))
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			handleError(w, r, fmt.Sprint(resp.StatusCode), err)
			return
		}
		defer resp.Body.Close()

		// Read the response body
		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}
	}

	// Unmarshal the response body into a slice of maps
	var packages []map[string]string
	if err := json.Unmarshal(respBody, &packages); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	for _, pkg := range packages {
		name := pkg["Name"]
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
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
			return
		}
		rating := string(body)
		pkg["Rating"] = rating
	}

	// Render the results page with the packages data
	tmpl, err := template.ParseFiles("templates/results.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	if err := tmpl.Execute(w, map[string]interface{}{"Packages": packages}); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func RenderHistory(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/history_search.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

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

	// Create a new HTTP client
	client := &http.Client{}

	// Make a GET request to the API
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
	if resp.StatusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(resp.StatusCode), err)
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	var packageHistories []models.ActionEntry
	err = json.Unmarshal(respBody, &packageHistories)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

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

	tmpl, err := template.ParseFiles("templates/history.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	if err := tmpl.Execute(w, packageHistoriesMaps); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func RenderDownload(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/download_search.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

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

	// Get the name
	id := r.FormValue("id")

	// Create a new HTTP client
	client := &http.Client{}

	// Make a GET request to the API
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
	if resp.StatusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(resp.StatusCode), err)
		return
	}
	defer resp.Body.Close()

	// Stream the response body and decode the JSON data in chunks
	var packageInfo map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&packageInfo); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	zip := packageInfo["data"].(map[string]interface{})["Content"].(string)
	name := packageInfo["metadata"].(map[string]interface{})["Name"].(string)
	zipSize, statusCode := utils.GetZipSize(zip)
	if statusCode != http.StatusOK {
		handleError(w, r, fmt.Sprint(statusCode), err)
		return
	}
	packageInfo["Size"] = zipSize

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
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
	rating := string(body)
	packageInfo["Rating"] = rating

	tmpl, err := template.ParseFiles("templates/download.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

	if err := tmpl.Execute(w, packageInfo); err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}
}

func RenderCreateReview(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/create_review.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

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

	bdy := map[string]string{"userName": userName, "packageName": packageName, "stars": stars, "review": review}

	bod, err := json.Marshal(bdy)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Call the API endpoint
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
		handleError(w, r, fmt.Sprint(resp.StatusCode), err)
		return
	}

	// Redirect the user
	handleSuccess(w, r)
}

func RenderDeleteReview(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/delete_review.html")
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusInternalServerError), err)
		return
	}

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

	bdy := map[string]string{"userName": userName, "packageName": packageName}

	bod, err := json.Marshal(bdy)
	if err != nil {
		handleError(w, r, fmt.Sprint(http.StatusBadRequest), err)
		return
	}

	// Call the API endpoint
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
		handleError(w, r, fmt.Sprint(resp.StatusCode), err)
		return
	}

	// Redirect the user
	handleSuccess(w, r)
}
