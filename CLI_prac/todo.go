package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Repo struct { //Structure that will recieve important information from REST API request
	URL         string `json:"html_url"`
	NetScore	int 
	RampUp		int	
	Correctness int
	BusFactor int
	ResponsiveMaintainer int
	License LName `json:"license"`
	Name string
}

type LName struct { //substructure to hold nested json fields
	Name string	`json:"name"`
}

type Repos []Repo

type respDataql struct { //type that stores data from graphql
	Repository struct {
		Issues struct {
			TotalCount int
		}
	}
}

func (r *Repos) Search(task string, resp *http.Response, NS int, RU int, C int, BF int, RM int) {

	var repo Repo
	json.NewDecoder(resp.Body).Decode(&repo) //decodes response and stores info in repo struct

	new_repo := Repo{ //setting values in repo struct, mostly hard coded for now.
		URL:         repo.URL,
		NetScore:	1,
		RampUp:		1,
		Correctness: 1,
		BusFactor: 1,
		ResponsiveMaintainer: 1,
		Name: repo.License.Name,
	}

	*r = append(*r, new_repo)

}

// func (t *Todos) Complete(index int) error {
// 	ls := *t
// 	if index <= 0 || index > len(ls){
// 		return errors.New("invalid index")
// 	}

// 	ls[index - 1].CompletedAt = time.Now()
// 	ls[index - 1].Done = true  //why index -1?

// 	return nil
// }

// func (t *Todos) Delete(index int) error { //in the future we could implement a sort of clear list function
// 	ls := *t
// 	if index <= 0 || index > len(ls){
// 		return errors.New("invalid index")
// 	}

// 	*t = append(ls[:index-1], ls[index:]...)

// 	return nil
// }


func (r *Repos) Load(filename string) error { //reads the json
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if len(file) == 0 {
		return err
	}

	err = json.Unmarshal([]byte(file), r)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repos) Store(filename string) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0644)
}

func (r *Repos) Print() {
	for _, repo := range *r {
		fmt.Printf("%s\n", repo.URL)
	}
}
