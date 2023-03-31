package models

type PackageInfo struct {
	Metadata Metadata    `json:"metadata,omitempty"`
	Data     PackageData `json:"data,omitempty"`
}
