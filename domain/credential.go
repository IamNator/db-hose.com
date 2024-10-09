package domain

type Credential struct {
	Key    string            `json:"key"`
	Values map[string]string `json:"values"`
}
