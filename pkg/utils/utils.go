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

// returns a metadata struct and a boolean indicating whether the package.json file was found. The last bool indicates whether the size of the package is too large
func extractPackageJsonFromZip(encodedZip string) (*PackageJson, bool, bool) {
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

	// Search for the package.json file in the zip archive
	var packageJson PackageJson
	found := false
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, "package.json") {
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

	// If the package.json file was not found, return an error (boolean false)
	// if !found {
	// 	return nil, errors.New("package.json not found in zip archive")
	// }

	return &packageJson, found, fileSize > MAX_FILE_SIZE
}

type RepoPackageJson struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func ExtractHomepageFromPackageJson(pkgJson PackageJson) string {
	var repourl string = ""
	if pkgJson.Homepage != "" {
		repourl = pkgJson.Homepage
	} else if str, ok := pkgJson.Repository.(string); ok {
		repourl = "https://github.com/" + str
		fmt.Println("Option 1", repourl)
		// fmt.Println(str)
	} else if repo, ok := pkgJson.Repository.(RepoPackageJson); ok {
		repourl = repo.URL
		fmt.Println("Option 2", repourl)

		// fmt.Println(repo.URL)
	} else if m, ok := pkgJson.Repository.(map[string]interface{}); ok {
		if url, ok := m["url"].(string); ok {
			repourl = url
			// fmt.Println("Option 3", repourl) This is the one that works
			repourl = strings.Replace(repourl, "http://", "https://", 1)
			repourl = strings.Replace(repourl, "git://", "https://", 1)
			// fmt.Println(url)
		}
	} else {
		return "" // GITHUB URL NOT FOUND
	}
	repourl = strings.TrimSuffix(repourl, ".git")

	return repourl
}

func GetStarsFromURL(gitURL string) float64 {
	re := regexp.MustCompile(`^https://github.com/[\w-]+/[\w-]+$`)
	if !re.MatchString(gitURL) {
		return 0.0
	}

	splitURL := strings.Split(gitURL, "/")
	user := splitURL[len(splitURL)-2]
	repo := splitURL[len(splitURL)-1]
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", user, repo)

	resp, err := http.Get(apiURL)
	if err != nil {
		return 0.0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0.0
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0.0
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0.0
	}

	stars := data["stargazers_count"].(float64) / 8000
	return stars
}

func CheckValidChars(input string) int {
	re := regexp.MustCompile(`^[\w-._~!$&'()*+,;=:@/?]+$`)
	if !re.MatchString(input) {
		return 0
	}
	return 1
}

// returns metadata, ifFound and if the package is too big
func ExtractMetadataFromZip(zipfile string) (models.Metadata, bool, bool) {
	pkgJson, found, tooBig := extractPackageJsonFromZip(zipfile)
	// fmt.Println(pkgJson.Name)
	// fmt.Println(pkgJson.Version)
	// fmt.Println(pkgJson.Repository)

	// TODO: parse different string variants of repository
	if !found {
		return models.Metadata{}, found, tooBig
	}

	// var repourl string
	// if pkgJson.Homepage != "" {
	// 	repourl = pkgJson.Homepage
	// } else if str, ok := pkgJson.Repository.(string); ok {
	// 	repourl = "https://github.com/" + str
	// 	fmt.Println("Option 1", repourl)
	// 	// fmt.Println(str)
	// } else if repo, ok := pkgJson.Repository.(RepoPackageJson); ok {
	// 	repourl = repo.URL
	// 	fmt.Println("Option 2", repourl)

	// 	// fmt.Println(repo.URL)
	// } else if m, ok := pkgJson.Repository.(map[string]interface{}); ok {
	// 	if url, ok := m["url"].(string); ok {
	// 		repourl = url
	// 		// fmt.Println("Option 3", repourl) This is the one that works
	// 		repourl = strings.Replace(repourl, "http://", "https://", 1)
	// 		repourl = strings.Replace(repourl, "git://", "https://", 1)
	// 		// fmt.Println(url)
	// 	}
	// } else {
	// 	return models.Metadata{}, false, false // GITHUB URL NOT FOUND
	// }

	// repourl = strings.TrimSuffix(repourl, ".git")
	repourl := ExtractHomepageFromPackageJson(*pkgJson)
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

func ExtractMetadataFromURL(url string) (models.Metadata, bool) {
	// Extract the repository owner and name from the URL
	parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
	if len(parts) < 2 {
		// fmt.Errorf("invalid Github URL: %s", url)
		return models.Metadata{}, false
	}
	owner, name := parts[0], parts[1]

	// Fetch the contents of the package.json file using Github's REST API
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/package.json", owner, name))
	if err != nil {
		// return PackageJSON{}, fmt.Errorf("error fetching package.json file: %v", err)
		return models.Metadata{}, false
	}
	defer resp.Body.Close()

	// Decode the base64-encoded content field
	var result struct {
		Content string `json:"content"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		// panic(err)
		return models.Metadata{}, false
	}

	content, err := base64.StdEncoding.DecodeString(result.Content)
	if err != nil {
		// panic(err)
		return models.Metadata{}, false
	}
	var packageJson PackageJson
	err = json.Unmarshal(content, &packageJson)
	if err != nil {
		// panic(err)
		return models.Metadata{}, false
	}
	// fmt.Println(packageJson.Name)
	// fmt.Println(packageJson.Version)
	// fmt.Println(packageJson.Repository)

	return models.Metadata{Name: packageJson.Name, Version: packageJson.Version, Repository: url, ID: "packageData_ID"}, true
}

func ExtractZipFromURL(url string) string {
	// Extract the repository owner and name from the URL
	parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
	if len(parts) < 2 {
		// fmt.Errorf("invalid Github URL: %s", url)
		// return "", false
	}
	owner, repo := parts[0], parts[1]

	// Send a GET request to the GitHub API endpoint
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball", owner, repo))
	if err != nil {
		// panic(err)
	}
	defer resp.Body.Close()

	// Read the contents of the downloaded zip file
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// panic(err)
	}

	// Encode the contents of the zip file using base64
	encoded := base64.StdEncoding.EncodeToString(contents)

	return encoded
}
