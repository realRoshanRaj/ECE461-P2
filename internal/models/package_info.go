package models

type PackageInfo struct {
	Metadata Metadata    `json:"metadata,omitempty" firestore:"metadata,omitempty"`
	Data     PackageData `json:"data,omitempty" firestore:"data,omitempty"`
}
