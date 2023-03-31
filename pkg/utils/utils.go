package utils

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"pkgmanager/internal/models"
	"strings"
)

type PackageJson struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Repository string `json:"repository"`
}

func extractPackageJsonFromZip(encodedZip string) (*PackageJson, error) {
	// Decode the base64-encoded string
	decoded, err := base64.StdEncoding.DecodeString(encodedZip)
	if err != nil {
		return nil, err
	}

	// Create a temporary file for the zip contents
	tempFile, err := ioutil.TempFile("", "tempzip-*.zip")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write the decoded zip contents to the temporary file
	_, err = tempFile.Write(decoded)
	if err != nil {
		return nil, err
	}

	// Open the zip file for reading
	reader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Search for the package.json file in the zip archive
	var packageJson PackageJson
	found := false
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, "package.json") {
			// Open the file from the zip archive
			zippedFile, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer zippedFile.Close()

			// Read the contents of the file into memory
			packageJsonBytes, err := ioutil.ReadAll(zippedFile)
			if err != nil {
				return nil, err
			}

			// Unmarshal the JSON into a struct
			err = json.Unmarshal(packageJsonBytes, &packageJson)
			if err != nil {
				return nil, err
			}

			found = true
			break
		}
	}

	// If the package.json file was not found, return an error
	if !found {
		return nil, errors.New("package.json not found in zip archive")
	}

	return &packageJson, nil
}

func ExtractMetadataFromZip(zipfile string) models.Metadata {
	pkgJson, _ := extractPackageJsonFromZip(zipfile)
	// fmt.Println(pkgJson.Name)
	// fmt.Println(pkgJson.Version)
	// fmt.Println(pkgJson.Repository)

	return models.Metadata{Name: pkgJson.Name, Version: pkgJson.Version, ID: "packageData_ID", Repository: pkgJson.Repository}
}

func ExtractMetadataFromURL(url string) models.Metadata {
	return models.Metadata{Name: "package_Name", Version: "package_Version", ID: "packageData_ID"}
}
