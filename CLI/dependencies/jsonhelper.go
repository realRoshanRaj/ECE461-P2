package dependencies

import (
	"encoding/json"
	"errors"
	"fmt"
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

func (r *Repos) Search(resp *http.Response) {

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

func (r *Repos) Load(filename string) error { //reads the json
	file, err := os.ReadFile(filename)
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

	return os.WriteFile(filename, data, 0666)
}


func (r *Repos) Print() {
	for _, repo := range *r {
		fmt.Printf("%s\n", repo.URL)
	}
}