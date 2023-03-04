package rest

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

// uses the npm registry api to get the data of a packaged
func getNPMData(pkgName string) NPMData {
	url := "https://registry.npmjs.org/" + pkgName
	response, err := http.Get(url)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	var responseObject NPMData
	json.Unmarshal(responseData, &responseObject)
	return responseObject
}

// takes in url and returns the package name
func parseNpmPackage(url string) string {
	npmLinkMatch := regexp.MustCompile(".*package/(.*)")
	return npmLinkMatch.FindStringSubmatch(url)[1]
}

// Takes in an NPM package name as the input and returns the raw https/ssh github url of the package
func GetGithubURL(pkgUrl string) string {
	pkgName := parseNpmPackage(pkgUrl)
	data := getNPMData(pkgName)
	return data.Repository.URL
}
