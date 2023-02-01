package main

import (
	"bufio"
	//"time"
	//"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"net/http"
	//"golang.org/x/oauth2"
	//"golang.org/x/oauth2/github"
	"context"
	//"log"
	"github.com/machinebox/graphql"

	"github.com/acestti/todo-app"

	"time"

	"github.com/joho/godotenv"
)

const (
	todoFile = ".todos.json"
)

func main() {

	token := os.Getenv("token")

	//add := flag.Bool("add", false, "add a new todo")
	search := flag.Bool("search", false, "search for repo")

	//complete := flag.Int("complete", 0, "mark an item as completed")
	//delete := flag.Int("delete", 0, "delete an item")
	list := flag.Bool("list", false, "list all todos")

	flag.Parse()
	todos := &todo.Repos{}

	if err := todos.Load(todoFile); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	switch { //cases for different flags
	case *search:
		input_URL, err := getInput(os.Stdin, flag.Args()...)
		input_parsed := strings.Split(input_URL, "/")

		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		client := &http.Client{}

		//GRAPH QL
		//scores [5]int 
		scores := graphql_func(input_parsed[3], input_parsed[4]) //better way than to copy array?
		fmt.Println(scores)

		//GRAPH QL END
		REST_api_link := "https://api.github.com/repos/" + input_parsed[3] + "/" + input_parsed[4] //converting github repo url to API url
		req, err := http.NewRequest("GET", REST_api_link, nil)
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Add("Authorization", token)

		// Make the GET request to the GitHub API
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("ERROR encountered /n/n")
			os.Exit(1)
		}
		defer resp.Body.Close()
		todos.Search(input_URL, resp, 1, 1, 1, 1, 1) //magic here

		err = todos.Store(todoFile)

	// case *complete > 0:
	// 	err := todos.Complete(*complete)
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, err.Error())
	// 		os.Exit(1)
	// 	}
	// 	err = todos.Store(todoFile)
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, err.Error())
	// 		os.Exit(1)
	// 	}

	// case *delete > 0:
	// 	err := todos.Delete(*delete)
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, err.Error())
	// 		os.Exit(1)
	// 	}
	// 	err = todos.Store(todoFile)
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, err.Error())
	// 		os.Exit(1)
	// 	}

	case *list:
		//graphql_func()
		// todos.Print()

	default:
		fmt.Fprintln(os.Stdout, "invalid Command")
		os.Exit(0)
	}
}

func getInput(r io.Reader, args ...string) (string, error) { //something for file piping, unnecessary for now
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	scanner := bufio.NewScanner(r)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}

	text := scanner.Text()

	if len(text) == 0 {
		return "", errors.New("Empty todo is not invalid")
	}

	return text, nil
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

func graphql_func(repo_owner string, repo_name string) []int { //seems to be working as long as token is stored in tokens.env
	// create a new client
	client := graphql.NewClient("https://api.github.com/graphql")


	scores := [5]int{0,0,0,0,0}
	// set the token for authentication
	godotenv.Load("tokens.env")
	token := os.Getenv("token")
	
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
	perc_PR := int(float64(respData1.Repository.PullRequests.TotalCount) * float64(0.4))
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
	difference := firstDate.Sub(secondDate).Hours()

	fmt.Println(difference)

	return scores[:]
}

