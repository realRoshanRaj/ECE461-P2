package models

type Metric struct {
	RepoURL              string  `json:"-"`
	NetScore             float64 `json:",omitempty"`
	BusFactor            float64 `json:",omitempty"`
	Correctness          float64 `json:",omitempty"`
	RampUp               float64 `json:",omitempty"`
	ResponsiveMaintainer float64 `json:",omitempty"`
	LicenseScore         float64 `json:",omitempty"`
	GoodPinningPractice  float64 `json:",omitempty"`
	PullRequest          float64 `json:",omitempty"`
}
