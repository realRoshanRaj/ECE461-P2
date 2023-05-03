package utils

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"pkgmanager/internal/models"
	"regexp"
	"strings"
)

const MAX_FILE_SIZE = 1000000

type PackageJson struct {
	Name       string      `json:"name"`
	Version    string      `json:"version"`
	Repository interface{} `json:"repository"`
	Homepage   string      `json:"homepage"`
}

type RepoPackageJson struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func GetZipSize(encodedZip string) (int, int) {
	// Decode the base64 string
	zip, err := base64.StdEncoding.DecodeString(encodedZip)
	if err != nil {
		return 0, http.StatusInternalServerError
	}

	// Calculate the size in KB
	sizeKB := float64(len(zip)) / 1024.0

	return int(sizeKB), http.StatusOK
}

func extractPackageJsonFromZip(encodedZip string) (*PackageJson, bool, bool) { // Returns the packageJson and a boolean indicating whether the package.json file was found. The last bool indicates whether the size of the package is too large
	// Decode the base64-encoded string
	decoded, err := base64.StdEncoding.DecodeString(encodedZip)
	if err != nil {
		return nil, false, false
	}

	// Create a temporary file for the zip contents
	tempFile, err := ioutil.TempFile("", "tempzip-*.zip")
	if err != nil {
		return nil, false, false
	}

	// Removes the zip file when we are done
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write the decoded zip contents to the temporary file
	_, err = tempFile.Write(decoded)
	if err != nil {
		return nil, false, false
	}
	// Get file information
	fileInfo, err := tempFile.Stat()
	if err != nil {
		return nil, false, false
	}

	// Get file size
	fileSize := fileInfo.Size()
	log.Println("File size: ", fileSize)

	// Open the zip file for reading
	reader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		return nil, false, false
	}
	defer reader.Close()

	// Search for the package.json file in the zip
	var packageJson PackageJson
	found := false
	for _, file := range reader.File {
		if strings.Count(file.Name, "/") == 1 && strings.HasSuffix(file.Name, "/package.json") { // Only match /package.json in root directory
			// Open the file from the zip archive
			zippedFile, err := file.Open()
			if err != nil {
				return nil, false, false
			}
			defer zippedFile.Close()

			// Read the contents of the file into memory
			packageJsonBytes, err := ioutil.ReadAll(zippedFile)
			if err != nil {
				return nil, false, false
			}

			// Unmarshal the JSON into a struct
			err = json.Unmarshal(packageJsonBytes, &packageJson)
			if err != nil {
				return nil, false, false
			}

			found = true
			break
		}
	}

	return &packageJson, found, fileSize > MAX_FILE_SIZE
}

func ExtractHomepageFromPackageJson(pkgJson PackageJson) string {
	var repourl string = ""
	// First check if the Homepage is not empty
	if pkgJson.Homepage != "" {
		repourl = pkgJson.Homepage
		if strings.HasPrefix(repourl, "https://github.com/") { // Checking if repourl is in the form of a GitHub URL
			repourl = strings.TrimSuffix(repourl, ".git")
			return repourl
		}
	}
	// Check if Repository key of packageJson is a string
	if str, ok := pkgJson.Repository.(string); ok {
		repourl = "https://github.com/" + str
		if strings.HasPrefix(repourl, "https://github.com/") { // Checking if repourl is in the form of a GitHub URL
			repourl = strings.TrimSuffix(repourl, ".git")
			return repourl
		}
	}
	// Check if Repository key of packageJson is Json
	if repo, ok := pkgJson.Repository.(RepoPackageJson); ok {
		repourl = repo.URL
		if strings.HasPrefix(repourl, "https://github.com/") { // Checking if repourl is in the form of a GitHub URL
			repourl = strings.TrimSuffix(repourl, ".git")
			return repourl
		}
	}
	// Check if Repository key of packageJson is a map
	if m, ok := pkgJson.Repository.(map[string]interface{}); ok {
		if url, ok := m["url"].(string); ok {
			repourl = url
			repourl = strings.Replace(repourl, "http://", "https://", 1) // Replace http with https
			repourl = strings.Replace(repourl, "git://", "https://", 1)  // Replace git with https
			if strings.HasPrefix(repourl, "https://github.com/") {       // Checking if repourl is in the form of a GitHub URL
				repourl = strings.TrimSuffix(repourl, ".git")
				return repourl
			}
		}
	}

	return "" // GITHUB URL NOT FOUND
}

func GetStarsFromURL(url string) float64 {
	// Get the owner and repo from the url
	found, owner, repo := getRepoFromURL(url)
	if !found {
		return 0.0
	}

	// Make the request
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	// Send the request
	resp, err := http.Get(apiURL)
	if err != nil {
		return 0.0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0.0
	}

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0.0
	}

	// Get the response into a map
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0.0
	}

	// Get the stars and scale the number of stars
	stars := data["stargazers_count"].(float64) / 8000
	return stars
}

func CheckValidChars(input string) int { // Some characters are disallowed by http If a disallowed character is passed in as a ID, this catches it and allows us to send bad request error
	re := regexp.MustCompile(`^[\w-._~!$&'()*+,;=:@/?]+$`)
	if !re.MatchString(input) {
		return 0
	}
	return 1
}

func ExtractMetadataFromZip(zipfile string) (models.Metadata, bool, bool) { // Returns the metadata and booleans representing if Found and if the package is too big
	pkgJson, found, tooBig := extractPackageJsonFromZip(zipfile)
	if !found {
		return models.Metadata{}, found, tooBig
	}
	// Gets the github URL from packageJson
	repourl := ExtractHomepageFromPackageJson(*pkgJson)
	// Returns a metadata struct with all of the required info
	return models.Metadata{Name: pkgJson.Name, Version: pkgJson.Version, ID: "packageData_ID", Repository: repourl}, found, tooBig
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

			// Convert the contents to a string
			readmeText := string(readmeBytes)
			return readmeText, http.StatusOK
		}
	}

	// If readme not found
	return "", http.StatusBadRequest
}

func GetReadmeTextFromGitHubURL(url string) (string, int) {
	// Get the owner and repo from the url
	found, owner, repo := getRepoFromURL(url)
	if !found {
		return "", http.StatusNotFound
	}
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/readme", owner, repo)

	// Make the request
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", http.StatusInternalServerError
	}

	// Read the response
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
		return "", http.StatusNotFound
	}

	// Get the download url to get the readme
	downloadURL := match[1]
	// Request the Readme
	resp, err = http.Get(downloadURL)
	if err != nil {
		return "", http.StatusInternalServerError
	}
	defer resp.Body.Close()

	// Read the response body
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", http.StatusInternalServerError
	}

	// Return README text as string
	return string(body), http.StatusOK
}

func ExtractMetadataFromURL(url string) (models.Metadata, bool) {
	// Get the owner and repo from the url
	found, owner, repo := getRepoFromURL(url)
	if !found {
		return models.Metadata{}, false
	}
	// Make the API request
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/package.json", owner, repo))
	if err != nil {
		return models.Metadata{}, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.Metadata{}, false
	}

	// Decode the base64-encoded content field
	var result struct {
		Content string `json:"content"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return models.Metadata{}, false
	}

	// Get the Content
	content, err := base64.StdEncoding.DecodeString(result.Content)
	if err != nil {
		return models.Metadata{}, false
	}
	// Get the packageJson
	var packageJson PackageJson
	err = json.Unmarshal(content, &packageJson)
	if err != nil {
		return models.Metadata{}, false
	}

	return models.Metadata{Name: packageJson.Name, Version: packageJson.Version, Repository: url, ID: "packageData_ID"}, true
}

func ExtractZipFromURL(url string) string {
	// Get the owner and repo from the url
	found, owner, repo := getRepoFromURL(url)
	if !found {
		return ""
	}

	// Send a GET request to the GitHub API endpoint
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball", owner, repo))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	// Read the contents of the downloaded zip file
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	// Encode the contents of the zip file using base64
	encoded := base64.StdEncoding.EncodeToString(contents)

	return encoded
}

func getRepoFromURL(gitURL string) (bool, string, string) {
	// Define the regex pattern to match GitHub repository URL
	regexPattern := `https?://github.com/([\w-]+)/([\w-]+)`

	// Compile the regex pattern
	regex := regexp.MustCompile(regexPattern)

	// Find the matches in the URL
	matches := regex.FindStringSubmatch(gitURL)

	// Make sure all parts of the url are present
	if len(matches) != 3 {
		return false, "", ""
	}

	// Fetch the contents of the package.json file using Github's REST API
	owner := matches[1]
	repo := matches[2]

	return true, owner, repo
}
