package controllers

import (
	"errors"
	"fmt"
	"github.com/olivercullimore/hikvision-anpr-alerts/app/models"
	"github.com/olivercullimore/hikvision-anpr-alerts/app/views"
	"html"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func NotFound(env *models.Env, w http.ResponseWriter, r *http.Request) {
	err := displayError(env, w, r, "404", "Oops! Page Not Found", "404 Page Not Found")
	if err != nil {
		env.Logger.Println(err)
	}
}

func MethodNotAllowed(env *models.Env, w http.ResponseWriter, r *http.Request) {
	err := displayError(env, w, r, "405", "Oops! Method Not Allowed", "405 Method Not Allowed")
	if err != nil {
		env.Logger.Println(err)
	}
}

// displayError will accept a ResponseWriter, Request, code, message and messagedetails and will output
// an error page in HTML format to the ResponseWriter.
func displayError(env *models.Env, w http.ResponseWriter, r *http.Request, code, message, messagedetails string) error {
	// Load error.html file if exists
	_, err := env.EmbedFS.Open("views/errors/error.html")
	errorOccurred := false
	if os.IsNotExist(err) {
		env.Logger.Printf("error file not found: %v\n", err)
		errorOccurred = true
	}
	if !errorOccurred {
		fileContent, err := env.EmbedFS.ReadFile("views/errors/error.html")
		if err != nil {
			env.Logger.Printf("error reading file: %v\n", err)
			errorOccurred = true
		}
		fileContentParsed := strings.Replace(string(fileContent), "{svgerror}", code, 1)
		fileContentParsed = strings.Replace(fileContentParsed, "{errormessage}", message, 1)
		fileContentParsed = strings.Replace(fileContentParsed, "{errormessagedetails}", messagedetails, 1)
		_, err = fmt.Fprint(w, fileContentParsed)
		if err != nil {
			return err
		}
	}
	// Output basic message if error.html doesn't exist
	if errorOccurred {
		_, err := fmt.Fprintf(w, "Error, sorry the page %q was not found.", html.EscapeString(r.URL.Path))
		if err != nil {
			return err
		}
	}
	return nil
}

// adminLog will accept an Environment, Request, logType and logDetails and will add
// an admin log to the database with the current session user id against it.
func adminLog(env *models.Env, r *http.Request, logType, logDetails string) error {
	// Get session
	session, err := env.SessionStore.Get(r, env.Config.SessionCookieName)
	if err != nil {
		return err
	}
	// Get user id
	userID, err := strconv.Atoi(fmt.Sprint(session.Values["userid"]))
	if err != nil {
		return err
	}
	// Add admin log to database
	adminLog := models.AdminLog{Type: logType, Details: logDetails, UserID: userID}
	_, err = adminLog.Add(env)
	if err != nil {
		return err
	}
	return nil
}

// getTheme will accept a Request and will return use light/dark theme.
func getTheme(r *http.Request) string {
	theme, err := r.Cookie("theme")
	if err != nil {
		return "light"
	}
	return strings.ReplaceAll(fmt.Sprint(theme), "theme=", "")
}

// getPageNumber will accept a Request and will return a page number.
func getPageNumber(r *http.Request) int {
	page := fmt.Sprint(r.URL.Query().Get("page"))
	pageNumber := 1
	if page != "" {
		pageNumber, _ = strconv.Atoi(page)
		if pageNumber < 1 {
			pageNumber = 1
		}
	}
	return pageNumber
}

// getPerPage will return a per page number.
func getPerPage(env *models.Env) int {
	perPage, _ := strconv.Atoi(fmt.Sprint(env.Config.PerPage))
	return perPage
}

// getPagination will accept a pageNumber and resCount and will return a pagination.
func getPagination(env *models.Env, pageNumber, resCount int) models.ListPagination {
	// Get per page
	perPage := getPerPage(env)

	// Get page number
	if pageNumber < 1 {
		pageNumber = 1
	}

	// Get page count
	pageCount := resCount / perPage
	if pageCount < 1 {
		pageCount = 1
	}

	// Set max page links
	maxPageLinks := 5 + pageNumber

	// Set pages
	var paginationPages []int
	if pageCount > 1 {
		for i := pageNumber; i <= pageCount; i++ {
			if i <= maxPageLinks || i == pageCount {
				if pageCount > maxPageLinks && i == pageCount {
					paginationPages = append(paginationPages, 0)
				}
				paginationPages = append(paginationPages, i)
			}
		}
	}

	// Set previous page
	paginationPrevious := pageNumber - 1
	if paginationPrevious < 1 || pageCount == 1 {
		paginationPrevious = 0
	}

	// Set next page
	paginationNext := pageNumber + 1
	if paginationNext > pageCount || pageCount == 1 {
		paginationNext = 0
	}

	return models.ListPagination{Current: pageNumber, Previous: paginationPrevious, Next: paginationNext, Pages: paginationPages}
}

func Login(env *models.Env, w http.ResponseWriter, r *http.Request) {
	var page = models.Page{Title: "Login", RequestURL: r.URL.String(), Theme: getTheme(r)}

	if r.Method == http.MethodPost {
		user := models.User{}

		// Parse form data ready for use
		err := r.ParseForm()
		if err != nil {
			err := displayError(env, w, r, "400", "Oops! Please try again later", "400 Bad Request")
			if err != nil {
				env.Logger.Println(err)
			}
			return
		}

		// Set values
		user.Email = fmt.Sprint(r.PostFormValue("email"))
		user.Password = fmt.Sprint(r.PostFormValue("password"))

		// Check required fields
		if user.Email != "" && user.Password != "" {
			userRes, res, err := user.CheckLogin(env)
			if err != nil {
				env.Logger.Println(err)
				if err == errors.New("user not found") {
					page.ErrorMessages = append(page.ErrorMessages, "Invalid login details")
				} else {
					err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
					if err != nil {
						env.Logger.Println(err)
					}
					return
				}
			}
			if res {
				// Get session
				session, err := env.SessionStore.Get(r, env.Config.SessionCookieName)
				if err != nil {
					env.Logger.Println(err)
					err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
					if err != nil {
						env.Logger.Println(err)
					}
					return
				}
				// Set user as authenticated
				session.Values["authenticated"] = true
				session.Values["userid"] = userRes.ID
				session.Values["useremail"] = userRes.Email
				loginRedirectURL := "/"
				if session.Values["loginredirecturl"].(string) != "" {
					loginRedirectURL = session.Values["loginredirecturl"].(string)
					session.Values["loginredirecturl"] = ""
				}
				err = session.Save(r, w)
				if err != nil {
					env.Logger.Println(err)
					err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
					if err != nil {
						env.Logger.Println(err)
					}
					return
				}
				// Redirect back to page
				http.Redirect(w, r, loginRedirectURL, 302)
			} else {
				page.ErrorMessages = append(page.ErrorMessages, "Invalid login details")
			}
		} else {
			page.ErrorMessages = append(page.ErrorMessages, "Missing required fields")
		}
	}
	views.Render(w, env, "login", http.StatusOK, page)
}

func Logout(env *models.Env, w http.ResponseWriter, r *http.Request) {
	// Get session
	session, err := env.SessionStore.Get(r, env.Config.SessionCookieName)
	if err != nil {
		env.Logger.Println(err)
		err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
		if err != nil {
			env.Logger.Println(err)
		}
		return
	}
	// Revoke user's authentication
	session.Values["authenticated"] = false
	err = session.Save(r, w)
	if err != nil {
		env.Logger.Println(err)
		err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
		if err != nil {
			env.Logger.Println(err)
		}
		return
	}
	// Redirect to login
	http.Redirect(w, r, "/login", 302)
}
