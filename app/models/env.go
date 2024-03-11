package models

import (
	"embed"
	"github.com/faabiosr/cachego"
	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
)

// Env struct
type Env struct {
	Config              Config
	Logger              *log.Logger
	DB                  *DB
	Cache               *cachego.Cache
	SessionStore        *sessions.CookieStore
	Validator           *validator.Validate
	ValidatorTranslator ut.Translator
	Templates           *template.Template
	EmbedFS             *embed.FS
}
