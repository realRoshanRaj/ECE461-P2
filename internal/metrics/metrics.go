package metrics

import "pkgmanager/internal/models"

func GenerateMetrics(url string) models.Metric {
	// TODO implement with part 1
	return models.Metric{}
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
