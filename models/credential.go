package models

type Credentials struct {
	Key    string            `json:"key"`
	Values map[string]string `json:"values"`
}
