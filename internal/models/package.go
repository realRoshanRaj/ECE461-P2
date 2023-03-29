package models

type Package struct {
	Metadata Metadata    `json:"metadata,omitempty"`
	Data     PackageData `json:"data,omitempty"`
}
