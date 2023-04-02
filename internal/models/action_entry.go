package models

type ActionEntry struct {
	Date     string   `json:"Date,omitempty" firestore:"Date,omitempty"`
	Metadata Metadata `json:"PackageMetadata,omitempty" firestore:"PackageMetadata,omitempty"`
	Action   string   `json:"Action,omitempty" firestore:"Action,omitempty"`
}
