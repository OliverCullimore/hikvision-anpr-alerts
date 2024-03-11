package models

// Config struct
type Config struct {
	HTTPHost          string
	HTTPPort          string
	ExternalURL       string
	SessionKey        string // SessionKey must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	SessionCookieName string
	PerPage           string
	SMTPHost          string
	SMTPPort          string
	SMTPUser          string
	SMTPPass          string
	SMTPAuth          string
	SMTPFrom          string
	DBFile            string
}
