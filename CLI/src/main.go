package main

import (
	dep "CLI/dependencies"
	"context"
	"math"
	"strconv"
	"time"

	// json "encoding/json"

	"fmt"

	// "io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"

	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	"github.com/machinebox/graphql"
)

const (
	testJson = "test.json"
)
var token string;
var repos *dep.Repos;

func init(){
	// Loads token into environment variables along with other things in the .env file
	godotenv.Load(".env")
	token = os.Getenv("GITHUB_TOKEN")
	repos = &dep.Repos{}

}
func main() {

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Printf("Please enter ./run help for help\n")
		os.Exit(0)
	}
	
	// Expects File path to be first arguement
	urlfile, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer urlfile.Close()

	// Read URLS from the file
	var urls []string
	scanner := bufio.NewScanner(urlfile)
	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		urls = append(urls, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// For each URL fetch data 
	for i := 0; i < len(urls); i++ {
		//if url is npm turn into github url
		convertUrl(&urls[i])

		// Gets HTTP response from Rest API
		repo_resp := getRepoResponse(urls[i]) // repository data
		contri_resp := getContributorResponse(urls[i]) //contributor data

		split_url := strings.Split(urls[i], "/")
		repo_owner := split_url[3];
		repo_name := split_url[4];

		// Gets Intermediate metric values from Graphql NOT FINAL SCORES
		metrics := graphql_func(repo_owner, repo_name, token)

		// Inserts the metrics into final function to do math on them and make a new struct out of them
		repos.Construct(repo_resp, contri_resp, metrics[1], metrics[2], metrics[3], metrics[4])
	}

	repos.Print()

	repos.Store(testJson)
}

// Converts npm url to github url
func convertUrl(url *string){
	if strings.HasPrefix(*url, "https://www.npmjs") {
		data, err := exec.Command("node", "giturl.js", *url).Output()
		if err != nil {
			log.Fatal(err)
		}
		*url = strings.TrimSuffix(string(data), "\n")
	}
}

func getRepoResponse(httpUrl string) *http.Response {
	client := &http.Client{}
	
	// Make sure the URL is to the repository main page
	link := strings.Split(httpUrl, "https://github.com/")
	REST_api_link := "https://api.github.com/repos/" + link[len(link)-1] //converting github repo url to API url
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
	
	/* Dumps the contents of the body of the request and the response
	*  into readable formats as in the html
	*/
	// LOGGING STUFF FOR DEBUGGING HTTP REQUESTS AND RESPONSES
	responseDump, err := httputil.DumpResponse(repo_resp, true)
	if err != nil {
		log.Fatalln(err)
	}
	// Here the 0666 is the same as chmod parameters in linux
	os.WriteFile("responseDumpRepo.log", responseDump, 0666)
	// This will DUMP your AUTHORIZATION token be careful! add to .gitignore if you haven't already
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Fatalln(err)
	}
	os.WriteFile("requestDumpRepo.log", requestDump, 0666)
	
	return repo_resp
}

func getContributorResponse(httpUrl string) *http.Response {
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
	responseDump, err := httputil.DumpResponse(repo_resp, true)
	if err != nil {
		log.Fatalln(err)
	}
	// Here the 0666 is the same as chmod parameters in linux
	os.WriteFile("responseDumpContributor.log", responseDump, 0666)
	// This will DUMP your AUTHORIZATION token be careful! add to .gitignore if you haven't already
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Fatalln(err)
	}
	os.WriteFile("requestDumpContributor.log", requestDump, 0666)

	return repo_resp
}

type respDataql1 struct { //type that stores data from graphql
	Repository struct {
		Issues struct {
			TotalCount int
		}
		PullRequests struct {
			TotalCount int
		}
		Upcase struct { //README.md
			Text string
		}
		Downcase struct { //readme.md
			Text string
		}
		Capcase struct { //Readme.md
			Text string
		}
		Expcase struct { //readme.markdown
			Text string
		}
		Commits struct {
			History struct {
				TotalCount int
			}
		}
	}
}

type respDataql2 struct {
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

func graphql_func(repo_owner string, repo_name string, token string) []float64 { //seems to be working as long as token is stored in tokens.env
	// create a new client
	client := graphql.NewClient("https://api.github.com/graphql")

	scores := [5]float64{0, 0, 0, 0, 0}

	// make a request
	req1 := graphql.NewRequest(`
	query { 
		repository(owner:"` + repo_owner + `", name:"` + repo_name + `") { 
			issues(states: OPEN) {
				totalCount
			}
			pullRequests(states: MERGED){
				totalCount
			}
			upcase: object(expression: "HEAD:README.md") {
				... on Blob {
					text
				}
			}
			downcase: object(expression: "HEAD:README.md") {
				... on Blob {
					text
				}
			}
			capcase: object(expression: "HEAD:Readme.md") {
				... on Blob {
					text
				}
			}
			expcase: object(expression: "HEAD:readme.markdown") {
				... on Blob {
					text
				}
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

	req1.Header.Add("Authorization", "Bearer "+token)
	var respData1 respDataql1
	if err := client.Run(context.Background(), req1, &respData1); err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Number of issues:", respData1.Repository.Downcase.Text)
	//40% of the last pull requests perhaps arbitrary number
	perc_PR1 := math.Min(20, float64(respData1.Repository.PullRequests.TotalCount)*float64(0.4))
	perc_PR := int(perc_PR1)
	//fmt.Println(perc_PR)

	req2 := graphql.NewRequest(`
	query {
		repository(owner:"` + repo_owner + `", name:"` + repo_name + `") { 
			issues(states: CLOSED) {
				totalCount
			}
			pullRequests (last: ` + strconv.Itoa(perc_PR) + `, states: MERGED) {
				nodes{
					createdAt
					mergedAt
				}
			}
		}
	}
	`)
	req2.Header.Add("Authorization", "Bearer "+token)

	var respData2 respDataql2
	if err := client.Run(context.Background(), req2, &respData2); err != nil {
		log.Fatal(err)
	}

	difference_sum := 0.0

	for i := 0; i < perc_PR; i++ {
		date1 := respData2.Repository.PullRequests.Nodes[i].MergedAt

		y1, err := strconv.Atoi(date1[0:3])
		if err != nil {
			return scores[:]
		}
		m1, err := strconv.Atoi(date1[5:6])
		if err != nil {
			return scores[:]
		}
		d1, err := strconv.Atoi(date1[8:9])
		if err != nil {
			return scores[:]
		}
		h1, err := strconv.Atoi(date1[11:12])
		if err != nil {
			return scores[:]
		}
		date2 := respData2.Repository.PullRequests.Nodes[i].CreatedAt
		y2, err := strconv.Atoi(date2[0:3])
		if err != nil {
			return scores[:]
		}

		m2, err := strconv.Atoi(date2[5:6])
		if err != nil {
			return scores[:]
		}
		d2, err := strconv.Atoi(date2[8:9])
		if err != nil {
			return scores[:]
		}
		h2, err := strconv.Atoi(date2[11:12])
		if err != nil {
			return scores[:]
		}

		firstDate := time.Date(y1, time.Month(m1), d1, h1, 0, 0, 0, time.UTC)
		secondDate := time.Date(y2, time.Month(m2), d2, h2, 0, 0, 0, time.UTC)
		difference_sum += math.Abs(firstDate.Sub(secondDate).Hours())
	}

	difference := difference_sum / float64(perc_PR)

	//time it takes to resolve, 3 days is the max, otherwise its a zero
	if difference > float64(72) {
		scores[4] = 0
	} else {
		scores[4] = dep.RoundFloat(1-(float64(difference)/float64(72)), 3)
	}

	//closed issues / total issues score of correctness
	scores[2] = dep.RoundFloat(float64(respData2.Repository.Issues.TotalCount)/(float64(respData1.Repository.Issues.TotalCount)+float64(respData2.Repository.Issues.TotalCount)), 3)

	//rampup... has readme
	if (respData1.Repository.Upcase.Text != "") || (respData1.Repository.Downcase.Text != "") ||
	 (respData1.Repository.Capcase.Text != "") || (respData1.Repository.Expcase.Text != ""){
		scores[1] = 1
	} else {
		scores[1] = 0
	}

	//will serve as denominator *NOT FINAL SCORE*
	scores[3] = float64(respData1.Repository.Commits.History.TotalCount)
	return scores[:]
}
