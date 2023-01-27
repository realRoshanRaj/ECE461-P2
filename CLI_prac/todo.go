package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	//"time"
)

type Repo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"html_url"`
}

// type item struct {
// 	Task	string
// 	Done	bool
// 	CreatedAt	time.Time
// 	CompletedAt time.Time
// }

type Repos []Repo

//type Todos []item

func (r *Repos) Search(task string, resp *http.Response) {
	// todo := item{
	// 	Task: task,
	// 	Done: false,
	// 	CreatedAt: time.Now(),
	// 	CompletedAt: time.Time{},
	// }

	var repo Repo
	json.NewDecoder(resp.Body).Decode(&repo)

	// fmt.Println("Name: ", repo.Name)
	// fmt.Println("Description: ", repo.Description)
	// fmt.Println("URL: ", repo.URL)

	new_repo := Repo{
		Name:        repo.Name,
		Description: repo.Description,
		URL:         repo.URL,
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

// func (t *Todos) Delete(index int) error {
// 	ls := *t
// 	if index <= 0 || index > len(ls){
// 		return errors.New("invalid index")
// 	}

// 	*t = append(ls[:index-1], ls[index:]...)

// 	return nil
// }

func (r *Repos) Load(filename string) error {
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

	err = json.Unmarshal(file, r)
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
		fmt.Printf("%s\n", repo.Name)
	}
}
