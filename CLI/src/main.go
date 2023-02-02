package main

import (
	dep "CLI/dependencies"
	// json "encoding/json"

	"flag"
	"fmt"

	// "io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"

	"os"
	"strings"
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

	getJsonFromHttpClient(args[0], string(token)) // using args[0] to test should be made sure is URL
	
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

func getJsonFromHttpClient(httpUrl string, token string){
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

	repos := &dep.Repos{}

	repos.Search(resp) // REALLY BAD NAME this doesn't search it decodes the new response and appends a repo struct into repos

	err = repos.Store(testJson)
	if err != nil {
		os.Exit(1)
	}
}
