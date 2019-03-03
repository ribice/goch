package goch

// User represents user entity
type User struct {
	UID         string `json:"uid"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Secret      string `json:"secret"`
}
