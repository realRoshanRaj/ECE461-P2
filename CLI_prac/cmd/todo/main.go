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
	"strings"

	"net/http"
	//"golang.org/x/oauth2" //
	//"golang.org/x/oauth2/github"
	"context"
	//"log"
	"github.com/machinebox/graphql"

	"github.com/acestti/todo-app"
)

const (
	todoFile = ".todos.json"
)

func main() {

	token := os.Getenv("token")

	//add := flag.Bool("add", false, "add a new todo")
	search := flag.Bool("search", false, "add a new todo")

	//complete := flag.Int("complete", 0, "mark an item as completed")
	//delete := flag.Int("delete", 0, "delete an item")
	list := flag.Bool("list", false, "list all todos")

	flag.Parse()
	todos := &todo.Repos{}

	if err := todos.Load(todoFile); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	switch {
	case *search:
		task, err := getInput(os.Stdin, flag.Args()...)

		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		
		client := &http.Client{}

		//GRAPHQL
		
		//GRAPHQL


		req, err := http.NewRequest("GET", "https://api.github.com/repos/TypeStrong/ts-node", nil)
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
		//fmt.Println(resp)
		defer resp.Body.Close()
		todos.Search(task, resp)

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
		graphql_func()
		todos.Print()

	default:
		fmt.Fprintln(os.Stdout, "invalid Command")
		os.Exit(0)
	}

	// Print the response

}

func getInput(r io.Reader, args ...string) (string, error) {
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

func graphql_func() {
	    // create a new client
		client := graphql.NewClient("https://api.github.com/graphql")

		// set the token for authentication
		
	
		// make a request
		req := graphql.NewRequest(`
			query { 
				repository(owner:"TypeStrong", name:"ts-node") { 
			 		issues(states:OPEN) {
						totalCount
			  		}
				}
		  	}
		`)

		req.Header.Add("Authorization", "")
	
		// run it and capture the response
		var respData struct {
			Repository struct {
				Issues struct {
					TotalCount int
				}
			}
		}
		if err := client.Run(context.Background(), req, &respData); err != nil {
			fmt.Println(err)
			return
		}
	
		fmt.Println("Number of issues:", respData.Repository.Issues.TotalCount)
}
