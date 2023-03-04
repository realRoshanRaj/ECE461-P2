package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

func GetGithubURL(pkgName string) string {
	data := getNPMData(pkgName)
	return data.Repository.URL
}
