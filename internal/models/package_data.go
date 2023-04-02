package models

type PackageData struct {
	Content   string `json:"Content,omitempty" firestore:"Content,omitempty"`
	URL       string `json:"URL,omitempty" firestore:"URL,omitempty"`
	JSProgram string `json:"JSProgram,omitempty"`
}
