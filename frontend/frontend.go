package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

// define a struct to hold the data for the template
type PageData struct {
	Title string
	Body  string
}

// Render function to render the template with the data
func RenderIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: "My Page",
		Body:  "Package Manager",
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RenderUpdate(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/update.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call the API endpoint
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, "http://localhost:8080/package/"+id, bytes.NewBuffer(bod))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "API error", resp.StatusCode)
		return
	}

	// Redirect the user back to the index page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func RenderCreate(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/create.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call the API endpoint
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/package/", bytes.NewBuffer(bod))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusCreated {
		http.Error(w, "API error", resp.StatusCode)
		return
	}

	// Redirect the user back to the index page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func RenderRemove(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/remove.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleRemove(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := r.FormValue("ID")
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, "http://localhost:8080/package/"+id, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "API error", resp.StatusCode)
		return
	}

	// Redirect the user back to the index page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func RenderReset(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/reset.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleReset(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, "http://localhost:8080/reset", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "API error", resp.StatusCode)
		return
	}

	// Not Redirecting but Handled
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func RenderSearch(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/search.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the value of the "name" field
	name := r.FormValue("name")
	version := r.FormValue("version")
	//id := r.FormValue("ID")
	//URL := r.FormValue("URL")
	//Content := r.FormValue("Content")
	//JSProgram := r.FormValue("JSProgram")
	bdy := []map[string]string{
		{"Version": version, "Name": name},
	}
	fmt.Println(bdy)
	bod, err := json.Marshal(bdy)
	fmt.Println(bod)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call the API endpoint
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/packages/", bytes.NewBuffer(bod))
	//print the response. not request

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprint(resp.StatusCode), resp.StatusCode)
		return
	}

	// Redirect the user back to the index page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
