package middleware

import (
	"github.com/olivercullimore/hikvision-anpr-alerts/app/models"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
	"net/http"
	"strings"
	"time"
)

// The Handler struct that takes a configured Env and a function matching
// our useful signature.
type AppHandler struct {
	Env     *models.Env
	Handler func(env *models.Env, w http.ResponseWriter, r *http.Request)
}

// ServeHTTP allows your type to satisfy the http.Handler interface.
func (ah *AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ah.Handler(ah.Env, w, r)
}

// Logging logs the incoming HTTP request & its duration.
func Logging(env *models.Env) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Recover from and log errors
			/*defer func() {
				if err := recover(); err != nil {
					//w.WriteHeader(http.StatusInternalServerError)
					env.Logger.Printf("%v %v\n", err, debug.Stack())
				}
			}()*/

			// Log request details
			if !strings.HasPrefix(r.URL.Path, "/static/") {
				env.Logger.Println(r.URL.Path)
			}
			// Call the next handler
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

// CORS sets the CORS headers.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		/*
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Access-Control-Request-Headers, Access-Control-Request-Method, Connection, Host, Origin, User-Agent, Referer, Cache-Control, X-header")
		*/
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Auth checks if the request is authenticated.
func Auth(env *models.Env) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/login" && !strings.HasPrefix(r.URL.Path, "/static/") {
				// Check staff is logged in
				session, err := env.SessionStore.Get(r, env.Config.SessionCookieName)
				if err != nil {
					env.Logger.Println(err)
					// Redirect to login page
					http.Redirect(w, r, "/login", 302)
					return
				}
				auth, ok := session.Values["authenticated"].(bool)
				if !ok || !auth {
					// Set login redirect URL
					session.Values["loginredirecturl"] = r.RequestURI
					err = session.Save(r, w)
					if err != nil {
						env.Logger.Println(err)
					}
					// Redirect to login page
					http.Redirect(w, r, "/login", 302)
					return
				}
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

// NewMemoryCache creates a new memory cache and returns
func NewMemoryCache(capacity int, ttl time.Duration) (*cache.Client, error) {
	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(capacity),
	)
	if err != nil {
		return nil, err
	}
	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(ttl),
		cache.ClientWithRefreshKey("opn"),
	)
	if err != nil {
		return nil, err
	}
	return cacheClient, nil
}
