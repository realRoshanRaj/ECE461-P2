package models

type Metadata struct {
	Name       string
	Version    string
	ID         string
	Repository string `json:"-"`
}
