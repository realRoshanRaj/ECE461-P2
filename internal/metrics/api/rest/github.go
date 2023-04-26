package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strings"

	"fmt"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type PackageJSON struct {
	Dependencies map[string]string `json:"dependencies"`
}

var token string

func init() {
	token = os.Getenv("GITHUB_TOKEN")
}

type PullRequest struct {
	Number githubv4.Int
}

type PageInfo struct {
	EndCursor   githubv4.String
	HasNextPage bool
}

type CommitResponse struct {
	Data struct {
		Repository struct {
			Ref struct {
				Target struct {
					History struct {
						TotalCount int `json:"totalCount"`
					} `json:"history"`
				} `json:"target"`
			} `json:"ref"`
		} `json:"repository"`
	} `json:"data"`
}

func GetNumCommits(owner string, repo string, token string, url string) (int, error) {
	defaultBranch := GetDefaultBranchName(url)
	query := fmt.Sprintf(`
	{
	  repository(owner: "%s", name: "%s") {
	    ref(qualifiedName: "%s") {
	      target {
	        ... on Commit {
	          history {
	            totalCount
	          }
	        }
	      }
	    }
	  }
	}
	`, owner, repo, defaultBranch)

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBufferString(fmt.Sprintf(`{"query": %q}`, query)))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var data CommitResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to decode response body: %s", string(body))
	}

	numCommits := data.Data.Repository.Ref.Target.History.TotalCount
	return numCommits, nil
}

type PullRequestConnection struct {
	TotalCount int
	Nodes      []struct {
		Commits struct {
			TotalCount int
		}
	} `graphql:"nodes"`
}

type Repository struct {
	DefaultBranchRef struct {
		Name string
	}
	PullRequests PullRequestConnection `graphql:"pullRequests(states: MERGED, baseRefName: $baseRefName, first: 100)"`
}

type Query struct {
	Repository Repository `graphql:"repository(owner: $owner, name: $name)"`
}

func GetCommitsInMergedPullRequests(owner string, name string, token string, url string) (int, error) {
	// Create a new authenticated GitHub client
	defaultBranch := GetDefaultBranchName(url)
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	// Define the variables for the GraphQL query
	variables := map[string]interface{}{
		"owner":       githubv4.String(owner),
		"name":        githubv4.String(name),
		"baseRefName": githubv4.String(defaultBranch),
	}

	// Execute the GraphQL query
	var query Query
	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		return 0, fmt.Errorf("error querying GitHub API: %v", err)
	}

	// Sum up the commit counts for each merged pull request
	totalCommits := 0
	for _, pr := range query.Repository.PullRequests.Nodes {
		totalCommits += pr.Commits.TotalCount
	}

	// Return the total commit count
	return totalCommits, nil
}

// TODO: change the log printf functions to new log
func GetPullRequestsResponse(httpUrl string) *http.Response {
	client := &http.Client{}

	// Make sure the URL is to the repository main page
	link := strings.Split(httpUrl, "https://github.com/")
	REST_api_link := "https://api.github.com/repos/" + link[len(link)-1] + "/pulls?state=closed" //converting github repo url to API url
	req, err := http.NewRequest(http.MethodGet, REST_api_link, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// Make the GET request to the GitHub API
	pr_resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	//defer pr_resp.Body.Close()

	return pr_resp
}

func GetPullRequestResponse(httpUrl string) *http.Response {
	client := &http.Client{}

	// Make sure the URL is to the repository main page
	req, err := http.NewRequest(http.MethodGet, httpUrl, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	// Make the GET request to the GitHub API
	pr_resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	return pr_resp
}

func GetCodeTabResponse(httpurl string) string {
	//client := &http.Client{}

	link := strings.Split(httpurl, "https://github.com/")
	Code_tab_link := "https://api.codetabs.com/v1/loc?github=" + link[len(link)-1] //converting github repo url to API url

	return Code_tab_link
}

func GetRepoResponse(httpUrl string) *http.Response {
	client := &http.Client{}

	// Make sure the URL is to the repository main page
	link := strings.Split(httpUrl, "https://github.com/")
	REST_api_link := "https://api.github.com/repos/" + link[len(link)-1] //converting github repo url to API url
	req, err := http.NewRequest(http.MethodGet, REST_api_link, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// Make the GET request to the GitH-ub API
	repo_resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer repo_resp.Body.Close()

	/* Dumps the contents of the body of the request and the response
	*  into readable formats as in the html
	 */
	// LOGGING STUFF FOR DEBUGGING HTTP REQUESTS AND RESPONSES
	_, err = httputil.DumpResponse(repo_resp, true)
	if err != nil {
		log.Fatalln(err)
	}
	// log.Printf("Response: %s", responseDump)
	// Here the 0666 is the same as chmod parameters in linux
	// os.WriteFile(log_file, responseDump, 0666) // Deprecated
	// This will DUMP your AUTHORIZATION token be careful! add to .gitignore if you haven't already
	_, err = httputil.DumpRequest(req, true)
	if err != nil {
		log.Fatalln(err)
	}
	// storeLog(log_file, requestDump, "Repo request dump\n", false)
	// storeLog(log_file, responseDump, "Repo response dump\n", false)

	return repo_resp
}

func GetContributorResponse(httpUrl string) *http.Response {
	client := &http.Client{}

	// Make sure the URL is the contributors page
	link := strings.Split(httpUrl, "https://github.com/")
	REST_api_link := "https://api.github.com/repos/" + link[len(link)-1] + "/contributors" //converting github repo url to API url
	// fmt.Println(REST_api_link)
	req, err := http.NewRequest(http.MethodGet, REST_api_link, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// Make the GET request to the GitHub API
	repo_resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer repo_resp.Body.Close()

	// LOGGING STUFF FOR DEBUGGING HTTP REQUESTS AND RESPONSES
	_, err = httputil.DumpResponse(repo_resp, true)
	if err != nil {
		log.Fatalln(err)
	}
	// log.Printf("Response: %s", responseDump)
	// Here the 0666 is the same as chmod parameters in linux
	// os.WriteFile(log_file, responseDump, 0666) // Deprecated
	// This will DUMP your AUTHORIZATION token be careful! add to .gitignore if you haven't already
	_, err = httputil.DumpRequest(req, true)
	if err != nil {
		log.Fatalln(err)
	}
	// log.Printf("Request: %s", _requestDump)
	// os.WriteFile("requestDumpContributor.log", requestDump, 0666) // Deprecate

	// storeLog(log_file, requestDump, "Contributor request dump\n", true)
	// storeLog(log_file, responseDump, "Contributor response dump\n", true)

	return repo_resp
}

func GetDefaultBranchName(httpUrl string) string {
	client := &http.Client{}

	// Make sure the URL is to the repository main page
	link := strings.Split(httpUrl, "https://github.com/")
	REST_api_link := "https://api.github.com/repos/" + link[len(link)-1] //converting github repo url to API url
	req, err := http.NewRequest(http.MethodGet, REST_api_link, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// Make the GET request to the GitH-ub API
	repo_resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer repo_resp.Body.Close()

	body, err := ioutil.ReadAll(repo_resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	contents := string(body)

	start_index := strings.Index(contents, `"default_branch"`) + len(`"default_branch"`)
	end_index := strings.Index(contents[start_index:], ",") + start_index
	defaultBranch := strings.TrimSpace(contents[start_index:end_index])
	defaultBranch = strings.Trim(defaultBranch, `"`)
	defaultBranch = strings.Trim(defaultBranch, `:`)
	defaultBranch = strings.Trim(defaultBranch, `"`)

	return defaultBranch

}

func GetVersionPinningResponse(httpUrl string) float64 {

	defaultBranch := GetDefaultBranchName(httpUrl)

	client := &http.Client{}

	// Make sure the URL is to the repository main page
	link := strings.Split(httpUrl, "https://github.com/")
	REST_api_link := "https://raw.githubusercontent.com/" + link[len(link)-1] + "/" + defaultBranch + "/" + "/package.json"
	req, err := http.NewRequest(http.MethodGet, REST_api_link, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// Make the GET request to the GitHub API
	repo_resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	contents, err := io.ReadAll(repo_resp.Body)

	if err != nil {
		log.Fatalln(err)
	}

	defer repo_resp.Body.Close()

	var package_data PackageJSON

	err = json.Unmarshal(contents, &package_data)

	if err != nil {
		log.Println(err)
	}

	if len(package_data.Dependencies) == 0 {
		return float64(1)
	}

	var total_counter float64
	var valid_counter float64

	total_counter = 0
	valid_counter = 0
	r := regexp.MustCompile(`^([0-9]+)(\.([0-9]+))*$`)
	for _, version := range package_data.Dependencies {

		if !(r.MatchString(string(version))) {
			total_counter += 1
			continue
		}

		valid_counter += 1
		total_counter += 1

	}

	return (valid_counter / total_counter)

}
