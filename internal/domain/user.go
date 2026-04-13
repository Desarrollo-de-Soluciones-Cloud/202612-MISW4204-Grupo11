package domain

// User representa un usuario sin secretos (respuestas HTTP).
type User struct {
	ID    int64    `json:"id"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}

// UserCredentials es lo mínimo para validar login (uso interno aplicación/repos).
type UserCredentials struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	Roles        []string
}
