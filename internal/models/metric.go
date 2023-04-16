package models

type Metric struct {
	RepoURL              string `json:"-"`
	NetScore             float64
	BusFactor            float64
	Correctness          float64
	RampUp               float64
	ResponsiveMaintainer float64
	LicenseScore         float64
	GoodPinningPractice  float64
	PullRequest          float64
}
