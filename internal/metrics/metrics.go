package metrics

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"pkgmanager/internal/metrics/api/graphql"
	"pkgmanager/internal/metrics/api/rest"
	"pkgmanager/internal/models"
	"pkgmanager/pkg/utils"
	"strings"
)

func GenerateMetrics(url string) models.Metric {
	// TODO implement with part 1
	token := os.Getenv("GITHUB_TOKEN")

	split_url := strings.Split(url, "/")
	repo_owner := split_url[3]
	repo_name := split_url[4]

	// repo_resp := rest.GetRepoResponse(url)          // repository data
	contri_resp := rest.GetContributorResponse(url) //contributor data
	totalPRs, err := rest.GetNumberOfMergedPRs(repo_owner, repo_name, token)
	if err != nil {
		log.Println(err)
	}

	numCommits, err := rest.GetNumCommits(repo_owner, repo_name, token)
	fraction := float64(totalPRs) / float64(numCommits)
	version_score := rest.GetVersionPinningResponse(url)

	metrics := graphql.Graphql_func(repo_owner, repo_name, token)
	// var repos *utils.Repos
	// url, net, bus_factor, correctness, rampup, responsiveness, license, pr, version := repos.Construct(repo_resp, contri_resp, metrics[0], metrics[1], metrics[2], metrics[3], metrics[4], fraction, version_score)
	packageRating := models.Metric{
		RepoURL:              url,
		BusFactor:            busFactorParse(contri_resp, metrics[3]),
		Correctness:          utils.RoundFloat(metrics[2], 2),
		RampUp:               utils.RoundFloat(metrics[1], 2),
		ResponsiveMaintainer: utils.RoundFloat(metrics[4], 2),
		LicenseScore:         utils.RoundFloat(metrics[0], 1),
		PullRequest:          utils.RoundFloat(fraction, 2),
		GoodPinningPractice:  utils.RoundFloat(version_score, 2)}
	packageRating.NetScore = utils.RoundFloat((packageRating.LicenseScore*(packageRating.Correctness+3*packageRating.ResponsiveMaintainer+packageRating.BusFactor+2*packageRating.RampUp))/7.0, 1)

	return packageRating
	// return models.Metric{}
}

func MeasureIngestibility(metrics models.Metric) bool {

	if metrics.NetScore < 0.5 {
		return false
	}

	if metrics.BusFactor < 0.5 {
		return false
	}

	if metrics.Correctness < 0.5 {
		return false
	}

	if metrics.RampUp < 0.5 {
		return false
	}

	if metrics.ResponsiveMaintainer < 0.5 {
		return false
	}

	if metrics.LicenseScore < 0.5 {
		return false
	}

	if metrics.GoodPinningPractice < 0.5 {
		return false
	}

	if metrics.PullRequest < 0.5 {
		return false
	}

	return true

}

func busFactorParse(contResp *http.Response, totalCommits float64) float64 {
	type Cont []struct { //best contributor
		Contributions int `json:"contributions"`
	}
	var cont Cont
	json.NewDecoder(contResp.Body).Decode(&cont) //decodes response and stores info in repo struct
	return utils.RoundFloat(1-(float64(cont[0].Contributions)/float64(totalCommits)), 2)
}
