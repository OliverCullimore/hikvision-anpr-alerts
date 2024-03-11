package app

import (
	"context"
	"embed"
	"encoding/hex"
	"fmt"
	filecache "github.com/faabiosr/cachego/file"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
	"github.com/matcornic/hermes/v2"
	envs "github.com/olivercullimore/go-utils/env"
	"github.com/olivercullimore/hikvision-anpr-alerts/app/models"
	"github.com/olivercullimore/hikvision-anpr-alerts/app/routes"
	"github.com/olivercullimore/hikvision-anpr-alerts/app/views"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

// Embed files
//
//go:embed all:views/errors all:views/static all:views/templates
var embedFS embed.FS

func Run() {
	// Initialize logger
	logger := log.New(os.Stdout, "app: ", log.LstdFlags|log.Lshortfile)

	// Load environment variables from file
	err := envs.Load(".env")
	if err != nil {
		logger.Println(err)
	}

	// Initialize config
	config := models.Config{}
	config.HTTPHost = checkConfig("HTTP_HOST", "0.0.0.0", "HTTP Host", "", logger)
	config.HTTPPort = checkConfig("HTTP_PORT", "80", "HTTP Port", "", logger)
	config.SessionCookieName = checkConfig("SESSION_COOKIE_NAME", "sessid", "Session Cookie Name", "", logger)
	config.ExternalURL = checkConfig("EXTERNAL_URL", "localhost", "External URL", "", logger)
	config.SessionKey = checkConfig("SESSION_KEY", "", "Session Key", "sessionkey", logger)
	config.PerPage = checkConfig("PER_PAGE", "40", "Per Page", "numeric", logger)
	config.SMTPHost = checkConfig("SMTP_HOST", "", "SMTP Host", "none", logger)
	config.SMTPPort = checkConfig("SMTP_PORT", "25", "SMTP Port", "numeric", logger)
	config.SMTPUser = checkConfig("SMTP_USER", "", "SMTP User", "none", logger)
	config.SMTPPass = checkConfig("SMTP_PASS", "", "SMTP Pass", "none", logger)
	config.SMTPAuth = checkConfig("SMTP_AUTH", "Unknown", "SMTP Auth", "none", logger)
	config.SMTPFrom = checkConfig("SMTP_FROM", "", "SMTP From", "none", logger)
	config.DBFile = checkConfig("DB_FILE", "./hikvision-anpr-alerts.db", "Database File", "none", logger)

	// Initialize cache store
	cache := filecache.New("/cache/")

	// Initialize database connection
	db := models.DB{}
	err = db.Init(config, cache, logger)
	if err != nil {
		logger.Printf("Error initializing database connection: %s\n", err)
		os.Exit(1)
	} else {
		logger.Println("Connected to database")
	}

	// Initialize session store
	sessionStore, err := InitSessionStore(config)
	if err != nil {
		logger.Printf("Error initializing session store: %s\n", err)
		os.Exit(1)
	} else {
		logger.Println("Initialized session store")
	}

	// Initialize validator
	validator, translator, err := InitValidator()
	if err != nil {
		logger.Printf("Error initializing validator: %s\n", err)
		os.Exit(1)
	} else {
		logger.Println("Initialized validator")
	}

	// Initialise env
	env := &models.Env{
		Config:              config,
		Logger:              logger,
		DB:                  &db,
		Cache:               &cache,
		SessionStore:        sessionStore,
		Validator:           validator,
		ValidatorTranslator: translator,
		EmbedFS:             &embedFS,
	}

	// Perform database migrations
	if err := db.Migrate(env); err != nil {
		env.Logger.Printf("Error performing database migrations: %s\n", err)
		os.Exit(1)
	} else {
		env.Logger.Println("Database migrations complete")
	}

	// Get cameras to connect to
	var camera models.Camera
	resCameras, resCamerasCount, err := camera.Find(env, "AND", []models.WhereFields{}, 0, 1)
	if err != nil {
		env.Logger.Println(err)
	}
	if resCamerasCount > 0 {
		// Create a wait group to wait for all connections to close
		var wg sync.WaitGroup
		// Connect to each camera
		for _, cam := range *resCameras {
			wg.Add(1)
			go func(c models.Camera, env *models.Env) {
				defer wg.Done()
				connectToCamera(c, env)
			}(cam, env)
		}
		// Listen for interrupts
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		signal.Notify(sigChan, os.Kill)
		sig := <-sigChan
		env.Logger.Println("Camera connections got signal:", sig)
		// Close camera connections gracefully
		wg.Wait()
	}

	// Load view templates
	err = views.Load(env)
	if err != nil {
		env.Logger.Printf("Error loading view templates: %s\n", err)
		os.Exit(1)
	} else {
		env.Logger.Println("Loaded view templates")
	}

	// Initialize router
	r := mux.NewRouter().StrictSlash(true)

	// Initialize routes
	routes.Initialize(r, env)

	// Initialize http server
	env.Logger.Println("Starting server at http://" + env.Config.HTTPHost + ":" + env.Config.HTTPPort)
	s := &http.Server{
		Addr:         ":" + env.Config.HTTPPort, // configure the bind address
		Handler:      r,                         // set the default handler
		ErrorLog:     env.Logger,                // set the logger for the server
		IdleTimeout:  120 * time.Second,         // max time to read request from the client
		ReadTimeout:  5 * time.Second,           // max time to write response to the client
		WriteTimeout: 10 * time.Second,          // max time for connections using TCP Keep-Alive
	}
	// Run http server
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			env.Logger.Printf("HTTP server error: %s\n", err)
			os.Exit(1)
		}
	}()
	// Listen for interrupts
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)
	sig := <-sigChan
	env.Logger.Println("HTTP server got signal:", sig)

	// Shutdown http server gracefully
	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = s.Shutdown(tc)
	if err != nil {
		env.Logger.Fatal(err)
	} else {
		env.Logger.Println("Shutdown Server")
	}
}

func checkConfig(envKey, defaultValue, name, validationType string, logger *log.Logger) string {
	valid := true
	// logger.Printf("n: %s e: %s d: %s", name, envValue, defaultValue)
	checkVal := ""
	envValue := envs.Get(envKey, "")
	if envValue != "" {
		checkVal = envValue
	}
	if checkVal == "" {
		checkVal = defaultValue
	}
	switch validationType {
	case "none":
		// no value check
	case "sessionkey":
		// key must not be empty and must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
		if checkVal == "" {
			valid = false
		}
		if len(checkVal) != 16 && len(checkVal) != 24 && len(checkVal) != 32 {
			valid = false
		}
	case "numeric":
		// value must be numeric
		_, err := strconv.Atoi(checkVal)
		if err != nil {
			valid = false
		}
	case "":
		// value must not be empty
		if checkVal == "" {
			valid = false
		}
	}
	if valid != true {
		// Generate a session key if the one used is invalid
		if validationType == "sessionkey" {
			logger.Printf("Please use the following value for the SESSION_KEY environment variable: %s\n", hex.EncodeToString(securecookie.GenerateRandomKey(16)))
		}
		logger.Fatalf("Invalid %s value", name)
	}
	return checkVal
}

func connectToCamera(cam models.Camera, env *models.Env) {
	// Create WebSocket URL
	wsURL := fmt.Sprintf("ws://%s/ISAPI/Event/notification/alertStream", cam.IPAddress)

	// Set up WebSocket connection
	u, err := url.Parse(wsURL)
	if err != nil {
		env.Logger.Printf("Error parsing WebSocket URL for %s: %v\n", cam.IPAddress, err)
		return
	}
	u.User = url.UserPassword(cam.Username, cam.Password)

	// Connect to WebSocket
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		env.Logger.Printf("Error connecting to WebSocket for %s: %v\n", cam.IPAddress, err)
		return
	}
	defer c.Close()

	// Handle incoming messages
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			env.Logger.Printf("Error reading message from %s: %v\n", cam.IPAddress, err)
			return
		}
		// Process the received event (e.g., log it)
		env.Logger.Printf("[%s] Received event: %s\n", cam.IPAddress, string(msg))
		// TODO: Send alert email if exists in number plate database
		// sendAlertEmail(numberPlate, env)
	}
}

func sendAlertEmail(numberplate *models.NumberPlate, env *models.Env) {
	if env.Config.SMTPFrom != "" {
		email := models.Email{
			To:      env.Config.SMTPFrom,
			Subject: "ANPR Alert",
			Body: hermes.Body{
				Name: "ANPR Alert",
				Intros: []string{
					"Test intro.",
				},
				Actions: []hermes.Action{
					{
						Instructions: "High priority customer",
						Button: hermes.Button{
							Color:     "#4285f4",
							TextColor: "#fff",
							Text:      "Test button",
							Link:      env.Config.ExternalURL,
						},
					},
				},
				Outros: []string{
					"Test outro.",
				},
			},
		}
		err := email.Send(env)
		if err != nil {
			env.Logger.Println(err)
			return
		}
	}
}
