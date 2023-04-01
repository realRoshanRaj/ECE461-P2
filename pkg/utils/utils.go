package utils

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
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

func extractPackageJsonFromZip(encodedZip string) (*PackageJson, bool) {
	// Decode the base64-encoded string
	decoded, err := base64.StdEncoding.DecodeString(encodedZip)
	if err != nil {
		return nil, false
	}

	// Create a temporary file for the zip contents
	tempFile, err := ioutil.TempFile("", "tempzip-*.zip")
	if err != nil {
		return nil, false
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write the decoded zip contents to the temporary file
	_, err = tempFile.Write(decoded)
	if err != nil {
		return nil, false
	}

	// Open the zip file for reading
	reader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		return nil, false
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
				return nil, false
			}
			defer zippedFile.Close()

			// Read the contents of the file into memory
			packageJsonBytes, err := ioutil.ReadAll(zippedFile)
			if err != nil {
				return nil, false
			}

			// Unmarshal the JSON into a struct
			err = json.Unmarshal(packageJsonBytes, &packageJson)
			if err != nil {
				return nil, false
			}

			found = true
			break
		}
	}

	// If the package.json file was not found, return an error (boolean false)
	// if !found {
	// 	return nil, errors.New("package.json not found in zip archive")
	// }

	return &packageJson, found
}

func ExtractMetadataFromZip(zipfile string) (models.Metadata, bool) {
	pkgJson, found := extractPackageJsonFromZip(zipfile)
	// fmt.Println(pkgJson.Name)
	// fmt.Println(pkgJson.Version)
	// fmt.Println(pkgJson.Repository)

	return models.Metadata{Name: pkgJson.Name, Version: pkgJson.Version, ID: "packageData_ID", Repository: pkgJson.Repository}, found
}

func ExtractMetadataFromURL(url string) models.Metadata {
	return models.Metadata{Name: "package_Name", Version: "package_Version", ID: "packageData_ID"}
}

func IsRatingQualified(metrics models.Metric) bool {
	// TODO: Check metric values against requirements
	return true
}
