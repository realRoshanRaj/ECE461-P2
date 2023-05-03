package metrics

/*
******************************************************************************************************************************
This file is currently not in use but can be used to extend the capabilities of the package manager to accept npm repositories
******************************************************************************************************************************
*/

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
)

type NPMData struct {
	ID          string `json:"_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Readme      string `json:"readme,omitempty"`
	Repository  struct {
		Type string `json:"type,omitempty"`
		URL  string `json:"url,omitempty"`
	} `json:"repository,omitempty"`
	License string `json:"license,omitempty"`
}

func getNPMData(pkgName string) NPMData { // Uses the npm registry api to get the data of a package
	// Creates the request
	url := "https://registry.npmjs.org/" + pkgName

	// Calls the API to get the package data
	response, err := http.Get(url)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	// Reads the response data
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	var responseObject NPMData
	err = json.Unmarshal(responseData, &responseObject)
	if err != nil {
		log.Fatal(err)
	}

	// Returns the response
	return responseObject
}

func parseNpmPackage(url string) string { // Takes in url and returns the package name
	npmLinkMatch := regexp.MustCompile(".*package/(.*)")
	return npmLinkMatch.FindStringSubmatch(url)[1]
}

func GetGithubURL(pkgUrl string) string { // Takes in an NPM package name as the input and returns the raw github url of the package
	pkgName := parseNpmPackage(pkgUrl)
	data := getNPMData(pkgName)
	return data.Repository.URL
}
