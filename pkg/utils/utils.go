package utils

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"pkgmanager/internal/models"
	"regexp"
	"strings"
)

type PackageJson struct {
	Name       string      `json:"name"`
	Version    string      `json:"version"`
	Repository interface{} `json:"repository"`
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

type RepoPackageJson struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func ExtractMetadataFromZip(zipfile string) (models.Metadata, bool) {
	pkgJson, found := extractPackageJsonFromZip(zipfile)
	// fmt.Println(pkgJson.Name)
	// fmt.Println(pkgJson.Version)
	// fmt.Println(pkgJson.Repository)

	// TODO: parse different string variants of repository

	var repourl string
	if str, ok := pkgJson.Repository.(string); ok {
		repourl = str
		// fmt.Println(str)
	} else if repo, ok := pkgJson.Repository.(RepoPackageJson); ok {
		repourl = repo.URL
		// fmt.Println(repo.URL)
	} else if m, ok := pkgJson.Repository.(map[string]interface{}); ok {
		if url, ok := m["url"].(string); ok {
			repourl = url
			// fmt.Println(url)
		}
	} else {
		return models.Metadata{}, false // GITHUB URL NOT FOUND
	}

	if found {
		return models.Metadata{Name: pkgJson.Name, Version: pkgJson.Version, ID: "packageData_ID", Repository: repourl}, found
	} else {
		return models.Metadata{}, found
	}
}

func GetReadmeFromZip(zipBase64 string) (string, int) {
	// Decode the base64-encoded zip file
	zipBytes, err := base64.StdEncoding.DecodeString(zipBase64)
	if err != nil {
		return "", http.StatusInternalServerError
	}

	// Create a reader from the zipBytes
	zipReader, err := zip.NewReader(strings.NewReader(string(zipBytes)), int64(len(zipBytes)))
	if err != nil {
		return "", http.StatusInternalServerError
	}

	// Define regular expression pattern for README file names
	pattern := "(R|r)(E|e)(A|a)(D|d)(M|m)(E|e)"
	regex := regexp.MustCompile(pattern)

	// Loop through each file in the zip archive
	for _, file := range zipReader.File {
		// Check if the file name matches the regular expression
		if regex.MatchString(strings.ToLower(file.Name)) {
			// Open the file
			zipFile, err := file.Open()
			if err != nil {
				return "", http.StatusInternalServerError
			}
			defer zipFile.Close()

			// Read the contents of the file
			readmeBytes, err := ioutil.ReadAll(zipFile)
			if err != nil {
				return "", http.StatusInternalServerError
			}

			// Convert the contents to string
			readmeText := string(readmeBytes)
			return readmeText, http.StatusOK
		}
	}

	return "", http.StatusBadRequest
}

func GetReadmeTextFromGitHubURL(url string) (string, int) {

	// Define the regex pattern to match GitHub repository URL
	regexPattern := `https?://github.com/([\w-]+)/([\w-]+)`

	// Compile the regex pattern
	regex := regexp.MustCompile(regexPattern)

	// Find the matches in the URL
	matches := regex.FindStringSubmatch(url)

	if len(matches) != 3 {
		return "", http.StatusInternalServerError
	}

	owner := matches[1]
	name := matches[2]

	repoURL := fmt.Sprintf("%s/%s", owner, name)

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/readme", repoURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", http.StatusInternalServerError
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", http.StatusInternalServerError
	}

	// Define the regex pattern to match the download URL in the API response
	regexPattern2 := `"download_url"\s*:\s*"([^"]+)"`

	// Compile the regex pattern
	regex2 := regexp.MustCompile(regexPattern2)

	// Find the matches in the API response
	match := regex2.FindStringSubmatch(string(body))

	if len(match) != 2 {
		return "", http.StatusInternalServerError
	}

	downloadURL := match[1]
	resp, err = http.Get(downloadURL)
	if err != nil {
		return "", http.StatusInternalServerError
	}
	defer resp.Body.Close()

	// Read response body
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", http.StatusInternalServerError
	}

	// Return README text as string
	return string(body), http.StatusOK
}

func ExtractMetadataFromURL(url string) models.Metadata {
	return models.Metadata{Name: "package_Name", Version: "package_Version", ID: "packageData_ID"}
}

func ExtractZipFromURL(url string) string {
	return "zipBase64"
}
