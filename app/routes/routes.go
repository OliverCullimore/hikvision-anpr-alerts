package routes

import (
	"github.com/gorilla/mux"
	"github.com/olivercullimore/hikvision-anpr-alerts/app/controllers"
	"github.com/olivercullimore/hikvision-anpr-alerts/app/middleware"
	"github.com/olivercullimore/hikvision-anpr-alerts/app/models"
	"io/fs"
	"net/http"
)

func Initialize(r *mux.Router, env *models.Env) {

	// Init memory caches
	/*
	   memoryCache500Miliseconds, err := middleware.NewMemoryCache(10000000, 500*time.Millisecond)
	     if err != nil {
	         env.Logger.Println(err)
	     }
	     memoryCache24Hours, err := middleware.NewMemoryCache(10000000, 24*time.Hour)
	     if err != nil {
	         env.Logger.Println(err)
	     }
	*/

	// Middleware and error handlers
	r.Use(middleware.Logging(env), middleware.Auth(env))
	r.NotFoundHandler = &middleware.AppHandler{env, controllers.NotFound}
	r.MethodNotAllowedHandler = &middleware.AppHandler{env, controllers.MethodNotAllowed}

	// Handle core routes
	r.Handle("/", &middleware.AppHandler{env, controllers.AdminNumberPlates}).Methods(http.MethodGet)
	r.Handle("/add", &middleware.AppHandler{env, controllers.AdminAddNumberPlate})
	r.Handle("/{id:[0-9]+}", &middleware.AppHandler{env, controllers.AdminEditNumberPlate})
	r.Handle("/{id:[0-9]+}/delete", &middleware.AppHandler{env, controllers.AdminDeleteNumberPlate})
	r.Handle("/cameras", &middleware.AppHandler{env, controllers.AdminCameras})
	r.Handle("/cameras/add", &middleware.AppHandler{env, controllers.AdminAddCamera})
	r.Handle("/cameras/{id:[0-9]+}", &middleware.AppHandler{env, controllers.AdminEditCamera})
	r.Handle("/cameras/{id:[0-9]+}/delete", &middleware.AppHandler{env, controllers.AdminDeleteCamera})
	r.Handle("/users", &middleware.AppHandler{env, controllers.AdminUsers})
	r.Handle("/users/add", &middleware.AppHandler{env, controllers.AdminAddUser})
	r.Handle("/users/{id:[0-9]+}", &middleware.AppHandler{env, controllers.AdminEditUser})
	r.Handle("/users/{id:[0-9]+}/delete", &middleware.AppHandler{env, controllers.AdminDeleteUser})
	r.Handle("/login", &middleware.AppHandler{env, controllers.Login}).Methods(http.MethodGet, http.MethodPost)
	r.Handle("/logout", &middleware.AppHandler{env, controllers.Logout}).Methods(http.MethodGet)

	// Handle static files
	staticFS, err := fs.Sub(env.EmbedFS, "views/static")
	if err != nil {
		env.Logger.Println(err)
	}
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
}
