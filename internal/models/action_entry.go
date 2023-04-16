package models

type ActionEntry struct {
	User     map[string]string `json:"User,omitempty" firestore:"User,omitempty"`
	Date     string            `json:"Date,omitempty" firestore:"Date,omitempty"`
	Metadata Metadata          `json:"PackageMetadata,omitempty" firestore:"PackageMetadata,omitempty"`
	Action   string            `json:"Action,omitempty" firestore:"Action,omitempty"`
}
