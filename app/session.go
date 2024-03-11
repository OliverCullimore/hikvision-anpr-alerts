package app

import (
	"github.com/gorilla/sessions"
	"github.com/olivercullimore/hikvision-anpr-alerts/app/models"
)

// InitSessionStore sets up the session cookie store.
func InitSessionStore(config models.Config) (*sessions.CookieStore, error) {
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key := []byte(config.SessionKey)
	return sessions.NewCookieStore(key), nil
}
