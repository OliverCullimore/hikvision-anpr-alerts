package views

import (
	"errors"
	"fmt"
	"github.com/olivercullimore/hikvision-anpr-alerts/app/models"
	"html/template"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Load will accept an environment and parses all templates into cache.
func Load(env *models.Env) error {
	templateFS, err := fs.Sub(env.EmbedFS, "views/templates")
	if err != nil {
		env.Logger.Println(err)
	}
	// Add functions and parse templates
	templates := template.New("")
	templates, err = template.New("").Funcs(template.FuncMap{
		"CurrentDate":     currentDate,
		"FormatDate":      formatDate,
		"FormatDate2":     formatDate2,
		"FormatDateTime":  formatDateTime,
		"FormatDateTime2": formatDateTime2,
		"FormatTime":      formatTime,
		"ClassName":       className,
	}).ParseFS(templateFS, "*.html")
	if err != nil {
		return err
	}
	// Set templates environment variable
	env.Templates = templates
	// Return
	return nil
}

// Render will accept a ResponseWriter, environment, template and page interface and writes the code
// and rendered template in HTML format to the ResponseWriter.
func Render(w http.ResponseWriter, env *models.Env, view string, code int, p models.Page) {
	// Check view template exists
	if env.Templates.Lookup(view) == nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		env.Logger.Println(errors.New("no such template " + view))
		return
	}
	// Render
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "text/html")
	// Execute template
	err := env.Templates.ExecuteTemplate(w, view, p)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		env.Logger.Println(err)
		return
	}
}

func currentDate(format string) template.HTML {
	// Return content
	return template.HTML(time.Now().Format(format))
}

func formatDate(t time.Time) template.HTML {
	// Return content
	return template.HTML(t.Format("02/01/2006"))
}

func formatDate2(t time.Time) template.HTML {
	// Return content
	return template.HTML(t.Format("2006-01-02"))
}

func formatDateTime(t time.Time) template.HTML {
	// Return content
	return template.HTML(t.Format("02/01/2006 15:04"))
}

func formatDateTime2(t time.Time) template.HTML {
	// Return content
	return template.HTML(t.Format("02/01/2006 15:04:05"))
}

func formatTime(t time.Time) template.HTML {
	// Return content
	return template.HTML(t.Format("15:04"))
}

func className(s string) template.HTML {
	// Return content
	return template.HTML(regexp.MustCompile(`[^a-zA-Z0-9\-]+`).ReplaceAllString(strings.ToLower(fmt.Sprint(s)), ""))
}
