package metrics

import (
	"math"
	"os"
	"pkgmanager/internal/models"
	"pkgmanager/pkg/utils"
	"strings"

	"github.com/apsystole/log"
)

func GenerateMetrics(url string) models.Metric {
	// Get the Github Token from the environment
	token := os.Getenv("GITHUB_TOKEN")

	// Get repo owner/name from the gitURL
	split_url := strings.Split(url, "/")
	repo_owner := split_url[3]
	repo_name := split_url[4]

	// Get scores/data from the graphql api calls [Correctness, Total Commits, Responsive Maintainer]
	metrics := GetMetricsFromGraphql(repo_owner, repo_name, token)

	// Get the Correctness Score
	correctness_score := metrics[0]

	// Get the bus factor (sending in Total Commits calculated above)
	total_commits := metrics[1]
	bus_factor := GetBusFactor(url, total_commits)

	responsive_score := metrics[2]

	// Get the number of commits in merges with pull requests for a repository
	commits_in_pr, err := GetCommitsInMergedPullRequests(repo_owner, repo_name, token, url)
	if err != nil {
		commits_in_pr = 0
		log.Println(err)
	}

	// Get the number of total commits for a repository
	numCommits, err := GetNumCommits(repo_owner, repo_name, token, url)
	if err != nil {
		log.Println(err)
		numCommits = commits_in_pr
	}

	// Get the fraction of commits that were merged with pull requests by total commits
	pullRequests := float64(commits_in_pr) / float64(numCommits)
	// Scales the Pull Request score
	pull_requests_score := float64(math.Sqrt(pullRequests)) + 0.05
	if pull_requests_score > 1.00 {
		pull_requests_score = 1.00
	}

	// Get the version pinning score
	version_score := GetVersionPinningResponse(url)

	// Get the RampUp score
	rampup_score := GetRampUp(url)

	// Get the License Score
	license_score := GetLicenseCompatibility(repo_owner, repo_name, url)

	// Calculate the Net Score
	net_score := float64((2*correctness_score)+(3*responsive_score)+bus_factor+(2*rampup_score)+pull_requests_score+version_score) / 10.0
	net_score = float64(license_score * net_score)

	// Set the Ratings data struct with the metrics
	packageRating := models.Metric{
		RepoURL:              url,
		BusFactor:            utils.RoundFloat(bus_factor, 2),
		Correctness:          utils.RoundFloat(correctness_score, 2),
		RampUp:               utils.RoundFloat(rampup_score, 2),
		ResponsiveMaintainer: utils.RoundFloat(responsive_score, 2),
		LicenseScore:         utils.RoundFloat(license_score, 1),
		PullRequest:          utils.RoundFloat(pull_requests_score, 2),
		GoodPinningPractice:  utils.RoundFloat(version_score, 2)}
	packageRating.NetScore = utils.RoundFloat(net_score, 1)
	log.Printf("Rate Package: %+v\n", packageRating)

	return packageRating
}

func MeasureIngestibility(metrics models.Metric) bool {
	// Checks if there are any metrics with a score less than 0.5 and returns false if so
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
