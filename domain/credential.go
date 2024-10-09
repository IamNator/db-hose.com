package domain

type Credential struct {
	ID        string            `json:"id"`
	Email     string            `json:"email"`
	Secret    map[string]string `json:"secret"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}
