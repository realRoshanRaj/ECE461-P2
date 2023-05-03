package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"pkgmanager/pkg/utils"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fmt"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/machinebox/graphql"
)

// Used to stote Contributor Data
type Cont []struct {
	Contributions int `json:"contributions"`
}

// Used to store the dependency versions from the PackageJSON
type PackageJSON struct {
	Dependencies map[string]string `json:"dependencies"`
}

// Initializes the Github Token
var token string

func init() {
	token = os.Getenv("GITHUB_TOKEN")
}

// Used to get the total number of commits
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

// Used to store the response from the first GraphQl call in the graphql function
type responseType1 struct {
	Repository struct {
		Issues struct {
			TotalCount int
		}
		PullRequests struct {
			TotalCount int
		}
		Commits struct {
			History struct {
				TotalCount int
			}
		}
	}
}

// Used to store the response from the second GraphQl call in the graphql function
type responseType2 struct {
	Repository struct {
		Issues struct {
			TotalCount int `json:"totalCount"`
		} `json:"issues"`
		PullRequests struct {
			Nodes []struct {
				CreatedAt string `json:"createdAt"`
				MergedAt  string `json:"mergedAt"`
			} `json:"nodes"`
		} `json:"pullRequests"`
	} `json:"repository"`
}

// Used as a part of the next struct
type PullRequestConnection struct {
	TotalCount int
	Nodes      []struct {
		Commits struct {
			TotalCount int
		}
	} `graphql:"nodes"`
}

// Used as a part of the Query struct
type Repository struct {
	DefaultBranchRef struct {
		Name string
	}
	PullRequests PullRequestConnection `graphql:"pullRequests(states: MERGED, baseRefName: $baseRefName, first: 100)"`
}

// Used to store Graphql queries
type Query struct {
	Repository Repository `graphql:"repository(owner: $owner, name: $name)"`
}

/*
***********************
Below are the functions
***********************
*/

func GetMetricsFromGraphql(repo_owner string, repo_name string, token string) []float64 { // Gets Correctness, Total commits, and Responsive Maintainer
	// Create a new client
	client := graphql.NewClient("https://api.github.com/graphql")
	scores := [3]float64{0, 0, 0} //[Correctness, Total commits, Responsive Maintainer]

	// Create a Graphql query
	request1 := graphql.NewRequest(`
	query {
		repository(owner:"` + repo_owner + `", name:"` + repo_name + `") {
			issues(states: OPEN) {
				totalCount
			}
			pullRequests(states: MERGED){
				totalCount
			}
			commits: object(expression: "HEAD") {
				... on Commit {
				  history {
					   totalCount
					}
				}
			}
		}
	}
	`)

	// Make the GraphQl request
	request1.Header.Add("Authorization", "Bearer "+token)
	var response1 responseType1
	if err := client.Run(context.Background(), request1, &response1); err != nil {
		log.Fatal(err)
	}

	// Number of previous PR's we want to check the responsiveness of
	var numPRs int
	totalPRs := response1.Repository.PullRequests.TotalCount
	if totalPRs > 50 {
		numPRs = 50
	} else {
		numPRs = totalPRs
	}

	// Creates a new graphQL query
	request2 := graphql.NewRequest(`
	query {
		repository(owner:"` + repo_owner + `", name:"` + repo_name + `") {
			issues(states: CLOSED) {
				totalCount
			}
			pullRequests (last: ` + strconv.Itoa(numPRs) + `, states: MERGED) {
				nodes{
					createdAt
					mergedAt
				}
			}
		}
	}
	`)
	request2.Header.Add("Authorization", "Bearer "+token)

	// Makes the request
	var response2 responseType2
	if err := client.Run(context.Background(), request2, &response2); err != nil {
		log.Fatal(err)
	}

	/*
		*********************************
		Calculating the Correctness Score
		*********************************
	*/
	closedIssues := float64(response2.Repository.Issues.TotalCount)
	openIssues := float64(response1.Repository.Issues.TotalCount)
	// Number of closed issues divided by total issues
	scores[0] = closedIssues / (openIssues + closedIssues)

	// Total Number of Commits, used to find Bus Factor in a later function
	scores[1] = float64(response1.Repository.Commits.History.TotalCount)

	/*
		*******************************************
		Calculating the Responsive Maintainer Score
		*******************************************
	*/

	// Sum of how long between when a pull request is created and merged
	differenceSum := 0.0

	// Used to get the date in this layout
	layout := "2006-01-02T15:04:05Z07:00"
	// Goes over the last 50 PRs
	for idx := 0; idx < numPRs; idx++ {
		tempDate := response2.Repository.PullRequests.Nodes[idx].MergedAt
		firstDate, err := time.Parse(layout, tempDate)
		if err != nil {
			return scores[:]
		}
		tempDate = response2.Repository.PullRequests.Nodes[idx].CreatedAt
		secondDate, err := time.Parse(layout, tempDate)
		if err != nil {
			return scores[:]
		}
		// Sums up the difference of time in hours
		differenceSum += math.Abs(firstDate.Sub(secondDate).Hours())
	}

	avgTime := differenceSum / float64(numPRs)
	if avgTime < float64(72) { // If the average time is less than 3 days, it gets a 1
		scores[2] = 1
	} else if avgTime > float64(750) { // If the average time is greater than 1 Month days, it gets a 0
		scores[2] = 0
	} else { // Otherwise it is scaled appropriately
		scores[2] = float64(math.Sqrt(1 - (float64(avgTime) / float64(750))))
	}

	return scores[:]
}

// List of all compatible licenses
var licenses = []string{
	"MIT", "Apache2", "BSD 3-Clause",
	"BSD 2-Clause", "ISC", "BSD Zero Clause",
	"Boost Software", "UPL", "Universal Permissive",
	"JSON", "Simple Public", "Copyfree Open Innovation",
	"Xerox", "Sendmail", "All-Permissive", "Artistic",
	"Berkely Database", "Modified BSD", "CeCILL", "Cryptix General",
	"Zope Public", "XFree86", "X11", "WxWidgets Library", "WTFPL",
	"WebM", "Unlicense", "StandardMLofNJ", "Ruby", "SGI Free Software",
	"Python", "Ruby", "Perl", "OpenLDAP", "Netscape Javascript", "NCSA",
	"Mozilla Public", "Intel Open Source"}

func GetLicenseCompatibility(repoOwner string, repoName string, url string) float64 {
	// Sends a request to the github api to get license
	api_url := "https://api.github.com/repos/" + repoOwner + "/" + repoName
	res, err := http.Get(api_url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Checks if the API request was ok
	if res.StatusCode != http.StatusOK {
		return 0
	}

	// Reads the response into a variable
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var repoInfo map[string]interface{}
	err = json.Unmarshal(body, &repoInfo)
	if err != nil {
		panic(err)
	}

	// Gets the License from the response
	var repoLicense string
	if repoInfo["license"] != nil && repoInfo["license"].(map[string]interface{})["key"] != nil {
		repoLicense = repoInfo["license"].(map[string]interface{})["key"].(string)
	}

	// Checks if the License was in the Github
	if repoLicense != "" {
		// Checks if the License is compatible
		for _, license := range licenses {
			matched, _ := regexp.MatchString("(?i)"+license, repoLicense)
			if matched {
				return 1.0
			}
		}
	} else {
		// If license not in the Github, it gets the readme
		readme, statusCode := utils.GetReadmeTextFromGitHubURL(url)
		if statusCode != http.StatusOK {
			return 0.0
		}
		// Checks if the License is compatible
		for _, license := range licenses {
			matched, _ := regexp.MatchString("(?i)"+license, readme)
			if matched {
				return 1.0
			}
		}
	}

	// Returned if license not found or if it's incompatible
	return 0.0
}

func GetRampUp(gitURL string) float64 {
	// Gets the Readme from the github
	readme, statusCode := utils.GetReadmeTextFromGitHubURL(gitURL)
	if statusCode != http.StatusOK {
		return 0.0
	}

	// Gets the length of the readme and scales it to get a score between 0 and 1
	readmeLen := float64(len(readme))
	if readmeLen/float64(5000) > 1 {
		return 1.0
	} else {
		return readmeLen / float64(5000)
	}
}

func GetNumCommits(owner string, repo string, token string, url string) (int, error) { // Getting the total number of commits to a repository
	// Get the default branch name
	defaultBranch := utils.GetDefaultBranchName(url, token)

	// Create the query
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

	// Make the GraphQL API call
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

	// Read the response from the API
	var data CommitResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to decode response body: %s", string(body))
	}

	// Get the total number of commits
	numCommits := data.Data.Repository.Ref.Target.History.TotalCount
	return numCommits, nil
}

func GetCommitsInMergedPullRequests(owner string, name string, token string, url string) (int, error) { // Get the number of commits in merges with pull requests for a repository
	// Get the default branch name
	defaultBranch := utils.GetDefaultBranchName(url, token)

	// Create a new authenticated GitHub client
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

func GetVersionPinningResponse(httpUrl string) float64 {
	// Get the default branch
	defaultBranch := utils.GetDefaultBranchName(httpUrl, token)

	// Create new client
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

	// Read the response into a variable
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
	total_counter = 0.0
	valid_counter = 0.0

	// RegEx to see if the version is valid orn not
	r := regexp.MustCompile(`^([0-9]+)(\.([0-9]+))*$`)

	// Goes through each version and checks against the RegEx
	for _, version := range package_data.Dependencies {
		if !(r.MatchString(string(version))) {
			total_counter += 1
			continue
		}
		valid_counter += 1
		total_counter += 1
	}

	// Pinning Score
	return (valid_counter / total_counter)

}

func GetBusFactor(httpUrl string, totalCommits float64) float64 {
	// Creates a new client
	client := &http.Client{}

	// Makes the request string
	link := strings.Split(httpUrl, "https://github.com/")
	REST_api_link := "https://api.github.com/repos/" + link[len(link)-1] + "/contributors" //converting github repo url to API url

	// Creates a request for the Github API
	req, err := http.NewRequest(http.MethodGet, REST_api_link, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// Makes the request to the GitHub API
	repo_resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer repo_resp.Body.Close()

	// Reads the response
	var cont Cont
	err = json.NewDecoder(repo_resp.Body).Decode(&cont)
	if err != nil {
		log.Println(err)
		return -1
	}

	// Gets one minus the ratio of the contributions by the most active contributor to total contributions
	busFactor := 1 - (float64(cont[0].Contributions) / float64(totalCommits))

	// Scales the bus factor appropriately
	busFactorScaled := busFactorScalingFunction(busFactor)
	return busFactorScaled
}

func busFactorScalingFunction(bf_score float64) float64 {
	// Scales the bus factor score as defined in Deliverable
	scaled := math.Log((math.E-1)*bf_score + 1)
	return float64(scaled)
}
