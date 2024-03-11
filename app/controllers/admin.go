package controllers

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/olivercullimore/hikvision-anpr-alerts/app/models"
	"github.com/olivercullimore/hikvision-anpr-alerts/app/views"
	"net/http"
	"strconv"
)

func AdminNumberPlates(env *models.Env, w http.ResponseWriter, r *http.Request) {
	var page = models.Page{Title: "Number Plates", RequestURL: r.URL.String(), Theme: getTheme(r)}

	list := models.List{}
	var listRowFields []models.ListRowField

	// Get page number
	pageNumber := getPageNumber(r)

	// Get all number plates
	var numberPlate models.NumberPlate
	resNumberPlates, resCount, err := numberPlate.Find(env, "AND", []models.WhereFields{}, getPerPage(env), pageNumber)
	if err != nil {
		env.Logger.Println(err)
		err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
		if err != nil {
			env.Logger.Println(err)
		}
		return
	}
	if resCount > 0 {
		listRowFields = append(listRowFields, models.ListRowField{Value: "Number Plate"})
		listRowFields = append(listRowFields, models.ListRowField{Value: "Name"})
		listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto", Value: "Actions"})
		listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto", Value: ""})
		list.Rows = append(list.Rows, models.ListRow{Fields: listRowFields})
		for _, resNumberPlate := range *resNumberPlates {
			var listRowFields []models.ListRowField
			listRowFields = append(listRowFields, models.ListRowField{Value: resNumberPlate.Plate})
			listRowFields = append(listRowFields, models.ListRowField{Value: resNumberPlate.Name})
			listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto field-padding-right", Type: "link", Class: "btn btn-icon btn-red", Link: fmt.Sprintf("/%v/delete", resNumberPlate.ID), Confirm: "Are you sure you want to delete this number plate?", Icon: "delete", Value: "Delete"})
			listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto field-padding-right", Type: "link", Class: "btn btn-icon btn-yellow", Link: fmt.Sprintf("/%v", resNumberPlate.ID), Icon: "pencil", Value: "Edit"})
			list.Rows = append(list.Rows, models.ListRow{Fields: listRowFields})
		}
		// Get pagination
		list.Pagination = getPagination(env, pageNumber, resCount)
	} else {
		listRowFields = []models.ListRowField{}
		listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto", Value: "No number plates found"})
		list.Rows = append(list.Rows, models.ListRow{Fields: listRowFields})
	}
	listRowFields = []models.ListRowField{}
	listRowFields = append(listRowFields, models.ListRowField{Type: "link", Class: "btn btn-icon btn-primary", Link: "/add", Icon: "plus", Value: "Add"})
	list.Rows = append(list.Rows, models.ListRow{Fields: listRowFields})

	page.View = list

	views.Render(w, env, "list", http.StatusOK, page)
}

func AdminAddNumberPlate(env *models.Env, w http.ResponseWriter, r *http.Request) {
	var page = models.Page{Title: "Add Number Plate", RequestURL: r.URL.String(), Theme: getTheme(r)}

	if r.Method == http.MethodPost {
		numberPlate := models.NumberPlate{}

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
		numberPlate.Plate = fmt.Sprint(r.Form["plate"][0])
		numberPlate.Name = fmt.Sprint(r.Form["name"][0])
		// Validate values
		err = env.Validator.Struct(numberPlate)
		if err != nil {
			for _, e := range err.(validator.ValidationErrors) {
				page.ErrorMessages = append(page.ErrorMessages, e.Translate(env.ValidatorTranslator))
			}
		}

		// Check for errors
		if len(page.ErrorMessages) == 0 {
			// Check required fields
			if numberPlate.Plate != "" {
				// Add number plate to database
				_, err = numberPlate.Add(env)
				if err != nil {
					env.Logger.Println(err)
					err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
					if err != nil {
						env.Logger.Println(err)
					}
					return
				} else {
					// Add admin log to database
					err = adminLog(env, r, "numberplate", fmt.Sprintf("Add number plate %s", numberPlate.Plate))
					if err != nil {
						env.Logger.Println(err)
					}
					// Redirect
					http.Redirect(w, r, "/", 302)
				}
			} else {
				page.ErrorMessages = append(page.ErrorMessages, "Missing required fields")
			}
		}

	}

	form := models.Form{CancelLink: "/"}
	form.Fields = append(form.Fields, models.FormField{Name: "plate", Title: "Number Plate *", Type: "text", Required: true, Placeholder: "Number Plate"})
	form.Fields = append(form.Fields, models.FormField{Name: "name", Title: "Name", Type: "text", Required: false, Placeholder: "Name"})
	form.SubmitName = "Save Changes"

	page.View = form

	views.Render(w, env, "form", http.StatusOK, page)
}

func AdminEditNumberPlate(env *models.Env, w http.ResponseWriter, r *http.Request) {
	var page = models.Page{Title: "Edit Number Plate", RequestURL: r.URL.String(), Theme: getTheme(r)}

	// Parse GET parameters ready for use
	vars := mux.Vars(r)

	numberPlateID, err := strconv.Atoi(fmt.Sprint(vars["id"]))
	if err != nil {
		env.Logger.Println(err)
		err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
		if err != nil {
			env.Logger.Println(err)
		}
		return
	}

	var numberPlate models.NumberPlate
	resNumberPlates, resNumberPlatesCount, err := numberPlate.Find(env, "AND", []models.WhereFields{{"id", "=", numberPlateID}}, 0, 1)
	if err != nil {
		env.Logger.Println(err)
		err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
		if err != nil {
			env.Logger.Println(err)
		}
		return
	}
	if resNumberPlatesCount > 0 {
		for _, resNumberPlate := range *resNumberPlates {
			numberPlate = resNumberPlate
		}
	}

	if r.Method == http.MethodPost {
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
		numberPlate.Plate = fmt.Sprint(r.Form["plate"][0])
		numberPlate.Name = fmt.Sprint(r.Form["name"][0])
		// Validate values
		err = env.Validator.Struct(numberPlate)
		if err != nil {
			for _, e := range err.(validator.ValidationErrors) {
				page.ErrorMessages = append(page.ErrorMessages, e.Translate(env.ValidatorTranslator))
			}
		}

		// Check for errors
		if len(page.ErrorMessages) == 0 {
			// Check required fields
			if numberPlate.Plate != "" {
				// Update number plate in database
				_, err = numberPlate.Update(env)
				if err != nil {
					env.Logger.Println(err)
					err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
					if err != nil {
						env.Logger.Println(err)
					}
					return
				} else {
					// Add admin log to database
					err = adminLog(env, r, "numberplate", fmt.Sprintf("Update number plate id %d", numberPlate.ID))
					if err != nil {
						env.Logger.Println(err)
					}
					// Redirect
					http.Redirect(w, r, "/", 302)
				}
			} else {
				page.ErrorMessages = append(page.ErrorMessages, "Missing required fields")
			}
		}

	}

	form := models.Form{CancelLink: "/"}
	form.Fields = append(form.Fields, models.FormField{Name: "plate", Title: "Number Plate *", Type: "text", Required: true, Placeholder: "Number Plate", Value: numberPlate.Plate})
	form.Fields = append(form.Fields, models.FormField{Name: "name", Title: "Name", Type: "text", Required: false, Placeholder: "Name", Value: numberPlate.Name})
	form.SubmitName = "Save Changes"

	page.View = form

	views.Render(w, env, "form", http.StatusOK, page)
}

func AdminDeleteNumberPlate(env *models.Env, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		numberPlate := models.NumberPlate{}

		// Parse GET parameters ready for use
		vars := mux.Vars(r)

		// Set values
		numberPlateID, err := strconv.Atoi(fmt.Sprint(vars["id"]))
		if err != nil {
			env.Logger.Println(err)
			err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
			if err != nil {
				env.Logger.Println(err)
			}
			return
		}

		// Delete number plate from database
		numberPlate.ID = numberPlateID
		_, err = numberPlate.Delete(env)
		if err != nil {
			env.Logger.Println(err)
			err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
			if err != nil {
				env.Logger.Println(err)
			}
			return
		}

		// Add admin log to database
		err = adminLog(env, r, "numberplate", fmt.Sprintf("Delete number plate id %d", numberPlate.ID))
		if err != nil {
			env.Logger.Println(err)
		}
	}

	// Redirect
	http.Redirect(w, r, "/", 302)
}

func AdminCameras(env *models.Env, w http.ResponseWriter, r *http.Request) {
	var page = models.Page{Title: "Cameras", RequestURL: r.URL.String(), Theme: getTheme(r)}

	list := models.List{}
	var listRowFields []models.ListRowField

	// Get page number
	pageNumber := getPageNumber(r)

	// Get all cameras
	var camera models.Camera
	resCameras, resCount, err := camera.Find(env, "AND", []models.WhereFields{}, getPerPage(env), pageNumber)
	if err != nil {
		env.Logger.Println(err)
		err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
		if err != nil {
			env.Logger.Println(err)
		}
		return
	}
	if resCount > 0 {
		listRowFields = append(listRowFields, models.ListRowField{Value: "Name"})
		listRowFields = append(listRowFields, models.ListRowField{Value: "IP Address"})
		listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto", Value: "Actions"})
		listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto", Value: ""})
		list.Rows = append(list.Rows, models.ListRow{Fields: listRowFields})
		for _, resCamera := range *resCameras {
			var listRowFields []models.ListRowField
			listRowFields = append(listRowFields, models.ListRowField{Value: resCamera.Name})
			listRowFields = append(listRowFields, models.ListRowField{Value: resCamera.IPAddress})
			listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto field-padding-right", Type: "link", Class: "btn btn-icon btn-red", Link: fmt.Sprintf("/cameras/%v/delete", resCamera.ID), Confirm: "Are you sure you want to delete this camera?", Icon: "delete", Value: "Delete"})
			listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto field-padding-right", Type: "link", Class: "btn btn-icon btn-yellow", Link: fmt.Sprintf("/cameras/%v", resCamera.ID), Icon: "pencil", Value: "Edit"})
			list.Rows = append(list.Rows, models.ListRow{Fields: listRowFields})
		}
		// Get pagination
		list.Pagination = getPagination(env, pageNumber, resCount)
	} else {
		listRowFields = []models.ListRowField{}
		listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto", Value: "No cameras found"})
		list.Rows = append(list.Rows, models.ListRow{Fields: listRowFields})
	}
	listRowFields = []models.ListRowField{}
	listRowFields = append(listRowFields, models.ListRowField{Type: "link", Class: "btn btn-icon btn-primary", Link: "/cameras/add", Icon: "plus", Value: "Add"})
	list.Rows = append(list.Rows, models.ListRow{Fields: listRowFields})

	page.View = list

	views.Render(w, env, "list", http.StatusOK, page)
}

func AdminAddCamera(env *models.Env, w http.ResponseWriter, r *http.Request) {
	var page = models.Page{Title: "Add Camera", RequestURL: r.URL.String(), Theme: getTheme(r)}

	if r.Method == http.MethodPost {
		camera := models.Camera{}

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
		camera.Name = fmt.Sprint(r.Form["name"][0])
		camera.IPAddress = fmt.Sprint(r.Form["ipaddress"][0])
		camera.Username = fmt.Sprint(r.Form["username"][0])
		camera.Password = fmt.Sprint(r.Form["password"][0])
		// Validate values
		err = env.Validator.Struct(camera)
		if err != nil {
			for _, e := range err.(validator.ValidationErrors) {
				page.ErrorMessages = append(page.ErrorMessages, e.Translate(env.ValidatorTranslator))
			}
		}

		// Check for errors
		if len(page.ErrorMessages) == 0 {
			// Check required fields
			if camera.IPAddress != "" && camera.Username != "" && camera.Password != "" {
				// Add camera to database
				_, err = camera.Add(env)
				if err != nil {
					env.Logger.Println(err)
					err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
					if err != nil {
						env.Logger.Println(err)
					}
					return
				} else {
					// Add admin log to database
					err = adminLog(env, r, "camera", fmt.Sprintf("Add camera %s", camera.IPAddress))
					if err != nil {
						env.Logger.Println(err)
					}
					// Redirect
					http.Redirect(w, r, "/cameras", 302)
				}
			} else {
				page.ErrorMessages = append(page.ErrorMessages, "Missing required fields")
			}
		}

	}

	form := models.Form{CancelLink: "/cameras"}
	form.Fields = append(form.Fields, models.FormField{Name: "name", Title: "Name", Type: "text", Required: false, Placeholder: "Name"})
	form.Fields = append(form.Fields, models.FormField{Name: "ipaddress", Title: "IP Address *", Type: "text", Required: true, Placeholder: "IP Address"})
	form.Fields = append(form.Fields, models.FormField{Name: "username", Title: "Username *", Type: "text", Required: true, Placeholder: "Username"})
	form.Fields = append(form.Fields, models.FormField{Name: "password", Title: "Password *", Type: "text", Required: true, Placeholder: "Password"})
	form.SubmitName = "Save Changes"

	page.View = form

	views.Render(w, env, "form", http.StatusOK, page)
}

func AdminEditCamera(env *models.Env, w http.ResponseWriter, r *http.Request) {
	var page = models.Page{Title: "Edit Camera", RequestURL: r.URL.String(), Theme: getTheme(r)}

	// Parse GET parameters ready for use
	vars := mux.Vars(r)

	cameraID, err := strconv.Atoi(fmt.Sprint(vars["id"]))
	if err != nil {
		env.Logger.Println(err)
		err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
		if err != nil {
			env.Logger.Println(err)
		}
		return
	}

	var camera models.Camera
	resCameras, resCamerasCount, err := camera.Find(env, "AND", []models.WhereFields{{"id", "=", cameraID}}, 0, 1)
	if err != nil {
		env.Logger.Println(err)
		err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
		if err != nil {
			env.Logger.Println(err)
		}
		return
	}
	if resCamerasCount > 0 {
		for _, resCamera := range *resCameras {
			camera = resCamera
		}
	}

	if r.Method == http.MethodPost {
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
		camera.Name = fmt.Sprint(r.Form["name"][0])
		camera.IPAddress = fmt.Sprint(r.Form["ipaddress"][0])
		camera.Username = fmt.Sprint(r.Form["username"][0])
		camera.Password = fmt.Sprint(r.Form["password"][0])
		// Validate values
		err = env.Validator.Struct(camera)
		if err != nil {
			for _, e := range err.(validator.ValidationErrors) {
				page.ErrorMessages = append(page.ErrorMessages, e.Translate(env.ValidatorTranslator))
			}
		}

		// Check for errors
		if len(page.ErrorMessages) == 0 {
			// Check required fields
			if camera.IPAddress != "" && camera.Username != "" && camera.Password != "" {
				// Update camera in database
				_, err = camera.Update(env)
				if err != nil {
					env.Logger.Println(err)
					err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
					if err != nil {
						env.Logger.Println(err)
					}
					return
				} else {
					// Add admin log to database
					err = adminLog(env, r, "camera", fmt.Sprintf("Update camera id %d", camera.ID))
					if err != nil {
						env.Logger.Println(err)
					}
					// Redirect
					http.Redirect(w, r, "/cameras", 302)
				}
			} else {
				page.ErrorMessages = append(page.ErrorMessages, "Missing required fields")
			}
		}

	}

	form := models.Form{CancelLink: "/cameras"}
	form.Fields = append(form.Fields, models.FormField{Name: "name", Title: "Name", Type: "text", Required: false, Placeholder: "Name", Value: camera.Name})
	form.Fields = append(form.Fields, models.FormField{Name: "ipaddress", Title: "IP Address *", Type: "text", Required: true, Placeholder: "IP Address", Value: camera.IPAddress})
	form.Fields = append(form.Fields, models.FormField{Name: "username", Title: "Username *", Type: "text", Required: true, Placeholder: "Username", Value: camera.Username})
	form.Fields = append(form.Fields, models.FormField{Name: "password", Title: "Password *", Type: "text", Required: true, Placeholder: "Password", Value: camera.Password})
	form.SubmitName = "Save Changes"

	page.View = form

	views.Render(w, env, "form", http.StatusOK, page)
}

func AdminDeleteCamera(env *models.Env, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		camera := models.Camera{}

		// Parse GET parameters ready for use
		vars := mux.Vars(r)

		// Set values
		cameraID, err := strconv.Atoi(fmt.Sprint(vars["id"]))
		if err != nil {
			env.Logger.Println(err)
			err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
			if err != nil {
				env.Logger.Println(err)
			}
			return
		}

		// Delete camera from database
		camera.ID = cameraID
		_, err = camera.Delete(env)
		if err != nil {
			env.Logger.Println(err)
			err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
			if err != nil {
				env.Logger.Println(err)
			}
			return
		}

		// Add admin log to database
		err = adminLog(env, r, "camera", fmt.Sprintf("Delete camera id %d", camera.ID))
		if err != nil {
			env.Logger.Println(err)
		}
	}

	// Redirect
	http.Redirect(w, r, "/cameras", 302)
}

func AdminUsers(env *models.Env, w http.ResponseWriter, r *http.Request) {
	var page = models.Page{Title: "Users", RequestURL: r.URL.String(), Theme: getTheme(r)}

	list := models.List{}
	var listRowFields []models.ListRowField

	// Get page number
	pageNumber := getPageNumber(r)

	// Get all users
	var user models.User
	resUsers, resCount, err := user.Find(env, "AND", []models.WhereFields{}, getPerPage(env), pageNumber)
	if err != nil {
		env.Logger.Println(err)
		err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
		if err != nil {
			env.Logger.Println(err)
		}
		return
	}
	if resCount > 0 {
		listRowFields = append(listRowFields, models.ListRowField{Value: "Name"})
		listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto", Value: "Actions"})
		listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto", Value: ""})
		list.Rows = append(list.Rows, models.ListRow{Fields: listRowFields})
		for _, resUser := range *resUsers {
			var listRowFields []models.ListRowField
			listRowFields = append(listRowFields, models.ListRowField{Value: resUser.Email})
			listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto field-padding-right", Type: "link", Class: "btn btn-icon btn-red", Link: fmt.Sprintf("/users/%v/delete", resUser.ID), Confirm: "Are you sure you want to delete this user?", Icon: "delete", Value: "Delete"})
			listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto field-padding-right", Type: "link", Class: "btn btn-icon btn-yellow", Link: fmt.Sprintf("/users/%v", resUser.ID), Icon: "pencil", Value: "Edit"})
			list.Rows = append(list.Rows, models.ListRow{Fields: listRowFields})
		}
		// Get pagination
		list.Pagination = getPagination(env, pageNumber, resCount)
	} else {
		listRowFields = []models.ListRowField{}
		listRowFields = append(listRowFields, models.ListRowField{FieldClass: " field-width-auto", Value: "No users found"})
		list.Rows = append(list.Rows, models.ListRow{Fields: listRowFields})
	}
	listRowFields = []models.ListRowField{}
	listRowFields = append(listRowFields, models.ListRowField{Type: "link", Class: "btn btn-icon btn-primary", Link: "/users/add", Icon: "plus", Value: "Add"})
	list.Rows = append(list.Rows, models.ListRow{Fields: listRowFields})

	page.View = list

	views.Render(w, env, "list", http.StatusOK, page)
}

func AdminAddUser(env *models.Env, w http.ResponseWriter, r *http.Request) {
	var page = models.Page{Title: "Add User", RequestURL: r.URL.String(), Theme: getTheme(r)}

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
		user.Email = fmt.Sprint(r.Form["email"][0])
		user.Password = fmt.Sprint(r.Form["password"][0])
		// Validate values
		err = env.Validator.Struct(user)
		if err != nil {
			for _, e := range err.(validator.ValidationErrors) {
				page.ErrorMessages = append(page.ErrorMessages, e.Translate(env.ValidatorTranslator))
			}
		}

		// Check for errors
		if len(page.ErrorMessages) == 0 {
			// Check required fields
			if user.Email != "" && user.Password != "" {
				// Add user to database
				_, err = user.Add(env)
				if err != nil {
					env.Logger.Println(err)
					if err == errors.New("user already exists") {
						page.ErrorMessages = append(page.ErrorMessages, "User already exists")
					} else {
						err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
						if err != nil {
							env.Logger.Println(err)
						}
						return
					}
				} else {
					// Add admin log to database
					err = adminLog(env, r, "user", fmt.Sprintf("Add user %s", user.Email))
					if err != nil {
						env.Logger.Println(err)
					}
					// Redirect
					http.Redirect(w, r, "/users", 302)
				}
			} else {
				page.ErrorMessages = append(page.ErrorMessages, "Missing required fields")
			}
		}
	}

	form := models.Form{CancelLink: "/users"}
	form.Fields = append(form.Fields, models.FormField{Name: "email", Title: "Email (required for role admin)", Type: "email", Required: false, Placeholder: "Email"})
	form.Fields = append(form.Fields, models.FormField{Name: "password", Title: "Password (required for role admin)", Type: "password", Required: false, Placeholder: "Password"})
	form.SubmitName = "Save Changes"

	page.View = form

	views.Render(w, env, "form", http.StatusOK, page)
}

func AdminEditUser(env *models.Env, w http.ResponseWriter, r *http.Request) {
	var page = models.Page{Title: "Edit User", RequestURL: r.URL.String(), Theme: getTheme(r)}

	// Parse GET parameters ready for use
	vars := mux.Vars(r)

	userID, err := strconv.Atoi(fmt.Sprint(vars["id"]))
	if err != nil {
		env.Logger.Println(err)
		err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
		if err != nil {
			env.Logger.Println(err)
		}
		return
	}

	user := models.User{}
	resUsers, resUsersCount, err := user.Find(env, "AND", []models.WhereFields{{"id", "=", userID}}, 0, 1)
	if err != nil {
		env.Logger.Println(err)
		err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
		if err != nil {
			env.Logger.Println(err)
		}
		return
	}
	if resUsersCount > 0 {
		for _, resUser := range *resUsers {
			user = resUser
		}
	}

	if r.Method == http.MethodPost {
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
		user.Email = fmt.Sprint(r.Form["email"][0])
		user.Password = fmt.Sprint(r.Form["password"][0])
		// Validate values
		err = env.Validator.Struct(user)
		if err != nil {
			for _, e := range err.(validator.ValidationErrors) {
				page.ErrorMessages = append(page.ErrorMessages, e.Translate(env.ValidatorTranslator))
			}
		}

		// Check for errors
		if len(page.ErrorMessages) == 0 {
			// Check required fields
			if user.Email != "" && (user.Password != "" || user.PasswordHash != "") {
				// Update user in database
				_, err = user.Update(env)
				if err != nil {
					env.Logger.Println(err)
					err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
					if err != nil {
						env.Logger.Println(err)
					}
					return
				} else {
					// Add admin log to database
					err = adminLog(env, r, "user", fmt.Sprintf("Update user id %d", user.ID))
					if err != nil {
						env.Logger.Println(err)
					}
					// Redirect
					http.Redirect(w, r, "/users", 302)
				}
			} else {
				page.ErrorMessages = append(page.ErrorMessages, "Missing required fields")
			}
		}

	}

	form := models.Form{CancelLink: "/users"}
	form.Fields = append(form.Fields, models.FormField{Name: "email", Title: "Email (required for role admin)", Type: "email", Required: false, Placeholder: "Email", Value: user.Email})
	form.Fields = append(form.Fields, models.FormField{Name: "password", Title: "Password (required for role admin)", Type: "password", Required: false, Placeholder: "Password", Value: ""})
	form.SubmitName = "Save Changes"

	page.View = form

	views.Render(w, env, "form", http.StatusOK, page)
}

func AdminDeleteUser(env *models.Env, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		user := models.User{}

		// Parse GET parameters ready for use
		vars := mux.Vars(r)

		// Set values
		userID, err := strconv.Atoi(fmt.Sprint(vars["id"]))
		if err != nil {
			env.Logger.Println(err)
			err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
			if err != nil {
				env.Logger.Println(err)
			}
			return
		}

		// Delete user from database
		user.ID = userID
		_, err = user.Delete(env)
		if err != nil {
			env.Logger.Println(err)
			err := displayError(env, w, r, "500", "Oops! Please try again later", "Internal Server Error")
			if err != nil {
				env.Logger.Println(err)
			}
			return
		}

		// Add admin log to database
		err = adminLog(env, r, "user", fmt.Sprintf("Delete user id %d", user.ID))
		if err != nil {
			env.Logger.Println(err)
		}
	}

	// Redirect
	http.Redirect(w, r, "/users", 302)
}
