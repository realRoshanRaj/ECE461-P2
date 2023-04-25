package metrics

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"pkgmanager/internal/metrics/api/graphql"
	"pkgmanager/internal/metrics/api/rest"
	"pkgmanager/internal/models"
	"pkgmanager/pkg/utils"
	"strings"

	"github.com/apsystole/log"
)

func GenerateMetrics(url string) models.Metric {
	// TODO implement with part 1
	token := os.Getenv("GITHUB_TOKEN")

	split_url := strings.Split(url, "/")
	repo_owner := split_url[3]
	repo_name := split_url[4]

	// repo_resp := rest.GetRepoResponse(url)          // repository data
	contri_resp := rest.GetContributorResponse(url) //contributor data
	commits_in_pr, err := rest.GetCommitsInMergedPullRequests(repo_owner, repo_name, token, url)
	if err != nil {
		log.Println(err)
	}

	numCommits, err := rest.GetNumCommits(repo_owner, repo_name, token, url)
	fraction := float64(commits_in_pr) / float64(numCommits)
	version_score := rest.GetVersionPinningResponse(url)

	metrics := graphql.Graphql_func(repo_owner, repo_name, token)
	// var repos *utils.Repos
	// url, net, bus_factor, correctness, rampup, responsiveness, license, pr, version := repos.Construct(repo_resp, contri_resp, metrics[0], metrics[1], metrics[2], metrics[3], metrics[4], fraction, version_score)
	packageRating := models.Metric{
		RepoURL:              url,
		BusFactor:            busFactorParse(contri_resp, metrics[3]),
		Correctness:          utils.RoundFloat(metrics[2], 2),
		RampUp:               utils.RoundFloat(metrics[1], 2),
		ResponsiveMaintainer: utils.RoundFloat(ResponsivenessScalingFunction(metrics[4]), 2),
		LicenseScore:         utils.RoundFloat(metrics[0], 1),
		PullRequest:          utils.RoundFloat(PullRequestScalingFunction(fraction), 2),
		GoodPinningPractice:  utils.RoundFloat(version_score, 2)}
	log.Printf("%+v\n", packageRating)
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
	return utils.RoundFloat(BusFactorScalingFunction(1-(float64(cont[0].Contributions)/float64(totalCommits))), 2)
}

func PullRequestScalingFunction(pr_score float64) float64 {

	return float64(math.Sqrt(pr_score))
}

func BusFactorScalingFunction(bf_score float64) float64 {

	scaled := math.Log((math.E-1)*bf_score + 1)
	return float64(scaled)
}

func ResponsivenessScalingFunction(rm_score float64) float64 {

	scaled := math.Log((math.E-1)*rm_score + 1)
	return float64(scaled)
}
