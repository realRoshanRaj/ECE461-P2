package main

import (
	dep "CLI/dependencies"
	"context"
	"math"
	"strconv"
	"time"

	// json "encoding/json"

	"flag"
	"fmt"

	// "io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"

	"os"
	"strings"

	"github.com/machinebox/graphql"
)
const (
	testJson = "test.json"
)

var input_URL string

func main() {
	
	args := os.Args[1:]
	if(len(args) == 0){
		fmt.Printf("Please enter ./run -help for help\n")
		os.Exit(0);
	}

	token, err := os.ReadFile(".env")
	if err != nil {
		log.Fatalln(err)
	}

	initFlags()

	resp := getHttpClient(args[0], string(token)) // using args[0] to test should be made sure is URL

	repos := &dep.Repos{}

	input_parsed := strings.Split(args[0], "/")
	metrics := graphql_func(input_parsed[3], input_parsed[4]) 
	
	repos.Search(args[0], resp, metrics[0], metrics[1], metrics[2], metrics[3], metrics[4])
	repos.Store(testJson)
}

func initFlags(){
	// TO-DO implement all the flags and their uses IF NEED BE

	input_URL = *(flag.String("search", "", "search for repo"))
	//list = flag.Bool("list", false, "list all todos")
	//add = flag.Bool("add", false, "add a new todo")
	//complete = flag.Int("complete", 0, "mark an item as completed")
	//delete = flag.Int("delete", 0, "delete an item")

	flag.Parse()
}

func getHttpClient(httpUrl string, token string) *http.Response {
	client := &http.Client{}

	link := strings.Split(httpUrl, "https://github.com/")
	REST_api_link := "https://api.github.com/repos/" + link[len(link)-1]//converting github repo url to API url
	req, err := http.NewRequest(http.MethodGet, REST_api_link, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Authorization", token)

	// Make the GET request to the GitHub API
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()


	/* Dumps the contents of the body of the request and the response 
	*  into readable formats as in the html
	*/
	responseDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatalln(err)
	}

	// Here the 0666 is the same as chmod parameters in linux
	os.WriteFile("responseDump.log", responseDump, 0666);

	// This will DUMP your AUTHORIZATION token be careful! add to .gitignore if you haven't already
	
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Fatalln(err)
	}
	os.WriteFile("requestDump.log", requestDump, 0666);

	return resp
}

type respDataql1 struct { //type that stores data from graphql
	Repository struct {
		Issues struct {
			TotalCount int
		}
		PullRequests struct {
			TotalCount int
		}
	}
}

type respDataql2 struct { //type that stores data from graphql
	Repository struct {
		Issues struct {
			TotalCount int
		}
		PullRequests struct{
			Nodes []struct{
				CreatedAt string
				MergedAt string
			}
		}
	}
}

func graphql_func(repo_owner string, repo_name string) []float64 { //seems to be working as long as token is stored in tokens.env
	// create a new client
	client := graphql.NewClient("https://api.github.com/graphql")

	
	scores := [5]float64{0,0,0,0,0}
	// set the token for authentication

	token1, err:= os.ReadFile(".env")
	token := string(token1)
	if err != nil {
		log.Fatal(err)
	}
	
	// make a request
	req1 := graphql.NewRequest(`
	query { 
		repository(owner:"`+repo_owner+`", name:"`+repo_name+`") { 
			issues(states: OPEN) {
				totalCount
			}
			pullRequests(states: MERGED){
				totalCount
			}
		}
	}
	`)
	
	req1.Header.Add("Authorization", "Bearer " + token)
	var respData1 respDataql1
	if err := client.Run(context.Background(), req1, &respData1); err != nil {
		fmt.Println(err)
		return scores[:]
	}
	//fmt.Println("Number of issues:", respData1.Repository.Issues.TotalCount)
	//40% of the last pull requests perhaps arbitrary number
	perc_PR1 := math.Min(20, float64(respData1.Repository.PullRequests.TotalCount) * float64(0.4))
	perc_PR := int(perc_PR1)
	//fmt.Println(perc_PR)
	
	req2 := graphql.NewRequest(`
	query {
		repository(owner:"`+repo_owner+`", name:"`+repo_name+`") { 
			issues(states: CLOSED) {
				totalCount
			}
			pullRequests (last: ` + strconv.Itoa(perc_PR)+ `, states: MERGED) {
				nodes{
					createdAt
					mergedAt
				}
			}
		}
	}
	`)
	req2.Header.Add("Authorization", "Bearer " + token)	
	
	var respData2 respDataql2
	if err := client.Run(context.Background(), req2, &respData2); err != nil {
		fmt.Println(err)
		return scores[:]
	}
	//fmt.Println(token)
	
	date1 := respData2.Repository.PullRequests.Nodes[0].MergedAt

	y1, err := strconv.Atoi(date1[0:3])
	if err != nil{
		return scores[:]
	}
	m1, err := strconv.Atoi(date1[5:6])
	if err != nil{
		return scores[:]
	}
	d1, err := strconv.Atoi(date1[8:9])
	if err != nil{
		return scores[:]
	}
	h1, err := strconv.Atoi(date1[11:12])
	if err != nil{
		fmt.Println("hello")
		return scores[:]
	}
	date2 := respData2.Repository.PullRequests.Nodes[0].CreatedAt
	y2, err := strconv.Atoi(date2[0:3])
	if err != nil{
		return scores[:]
	}

	m2, err := strconv.Atoi(date2[5:6])
	if err != nil{
		return scores[:]
	}
	d2, err := strconv.Atoi(date2[8:9])
	if err != nil{
		return scores[:]
	}
	h2, err := strconv.Atoi(date2[11:12])
	if err != nil{
		return scores[:]
	}

	firstDate := time.Date(y1, time.Month(m1), d1, h1, 0, 0, 0, time.UTC)
    secondDate := time.Date(y2, time.Month(m2), d2, h2, 0, 0, 0, time.UTC)
	difference := math.Abs(firstDate.Sub(secondDate).Hours())

	//time it takes to resolve, 3 days is the max, otherwise its a zero
	if difference > float64(72){
		scores[4] = 0
	}else{
		scores[4] = roundFloat(1 - (float64(difference) / float64(72)), 3)
		fmt.Printf("differenece: %f\n",difference)
	}

	//closed issues / total issues score of correctness
	scores[2] = roundFloat(float64(respData2.Repository.Issues.TotalCount) / (float64(respData1.Repository.Issues.TotalCount) + float64(respData2.Repository.Issues.TotalCount)), 3)
	
	fmt.Println(scores)
	return scores[:]
}

func roundFloat(val float64, precision uint) float64 {
    ratio := math.Pow(10, float64(precision))
    return math.Round(val*ratio) / ratio
}