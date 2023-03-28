package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	// These are dependencies must be installed with go get
	// nd "github.com/scizorman/go-ndjson"
)

type Cont []struct { //best contributor
	Contributions int `json:"contributions"`
}

// type NCont struct { //nested info about contributor
// 	Contributions int `json:"contributions"`
// 	Id	int `json:"id"`
// }

type Repo struct { //Structure that will recieve important information from REST API request
	URL                         string `json:"URL"`
	NET_SCORE                   float64
	RAMP_UP_SCORE               float64
	CORRECTNESS_SCORE           float64
	BUS_FACTOR_SCORE            float64
	RESPONSIVE_MAINTAINER_SCORE float64
	LICENSE_SCORE               float64
	CODE_QUALITY_SCORE          float64
	// License              LName `json:"license"`
	// Name string
}

type LName struct { //substructure to hold nested json fields
	Name string `json:"name"`
}

type Repos []Repo

func (r *Repos) Construct(resp *http.Response, resp1 *http.Response, LS float64, RU float64, C float64, totalCommits float64, RM float64) {

	var repo Repo
	json.NewDecoder(resp.Body).Decode(&repo) //decodes response and stores info in repo struct
	//fmt.Println(repo.License.Name)

	var cont Cont
	json.NewDecoder(resp1.Body).Decode(&cont) //decodes response and stores info in repo struct
	//fmt.Println(cont[0].Contributions)

	if (repo == Repo{}) {
		log.Fatal("repo struct empty, check http response")
	}
	if cont == nil {
		log.Fatal("cont struct empty, check http response")
	}

	new_repo := Repo{ //setting values in repo struct, mostly hard coded for now.
		URL:                         repo.URL,
		RAMP_UP_SCORE:               RoundFloat(RU, 1),
		CORRECTNESS_SCORE:           RoundFloat(C, 1),
		BUS_FACTOR_SCORE:            RoundFloat(1-(float64(cont[0].Contributions)/totalCommits), 1),
		RESPONSIVE_MAINTAINER_SCORE: RoundFloat(RM, 1),
		LICENSE_SCORE:               RoundFloat(LS, 1),
		CODE_QUALITY_SCORE:          RoundFloat(0, 1),
		// License:              repo.License,
	}

	// var LicenseComp float64
	// if (new_repo.License.Name != "") {
	// 	LicenseComp = 1
	// } else {
	// 	LicenseComp = 0
	// }
	new_repo.NET_SCORE = RoundFloat((new_repo.LICENSE_SCORE*(new_repo.CORRECTNESS_SCORE+3*new_repo.RESPONSIVE_MAINTAINER_SCORE+new_repo.BUS_FACTOR_SCORE+2*new_repo.RAMP_UP_SCORE))/7.0, 1)
	// new_repo.LicenseScore = LS
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

	// This would be needed if we needed to append to file instead
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := os.Truncate(filename, 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}
	// This would be used if we needed to overwrite a file instead
	// if _, err := f.WriteString(string(ndata)); err != nil {
	// 	log.Fatal(err)
	// }

	for _, repo := range *r {
		data, err := json.Marshal(repo)
		if err != nil {
			return err
		}
		if _, err := f.Write(data); err != nil {
			log.Fatal(err)
		}
		f.WriteString("\n")
	}

	// os.WriteFile(filename,data, 0644);
	return err
}

func (r *Repos) Print() error {

	// fmt.Printf("Format\n")
	// fmt.Printf("https://host.com/url/to/repository\n")
	// fmt.Printf("NetScore    RampUp    Correctness    BusFactor    ResponsiveMaintainer    license\n")
	// for _, repo := range *r {
	// 	fmt.Printf("%s\n", repo.URL)
	// 	fmt.Printf("%.3f	%.3f	%.3f	%.3f	%.3f	%.3f\n", repo.NetScore, repo.RampUp, repo.Correctness, repo.BusFactor, repo.ResponsiveMaintainer, repo.LicenseScore)
	// }
	for _, repo := range *r {
		data, err := json.Marshal(repo)
		if err != nil {
			return err
		}
		fmt.Printf(string(data))
		fmt.Printf("\n")
	}

	return nil
}

func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
