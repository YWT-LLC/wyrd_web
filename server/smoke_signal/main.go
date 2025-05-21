package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/rs/cors"
	"github.com/wader/gormstore/v2"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// RootHandler renders and writes out the API schema definition
func RootHandler(w http.ResponseWriter, r *http.Request) {
	var body interface{}
	file, err := os.Open("open-api.yaml")
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
	}

	dec := yaml.NewDecoder(file)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	err = dec.Decode(&body)
	if err != nil {
		http.Error(w, "Server error", http.StatusServiceUnavailable)
	}
	body = utils.Convert(body)
	err = enc.Encode(body)
	if err != nil {
		http.Error(w, "Server error", http.StatusServiceUnavailable)
	}
}

func mainRouter() *mux.Router {
	var corsAllowedOrigins []string
	jsonOrigins, ok := os.LookupEnv("CORS_ALLOW_ORIGIN")
	if ok {
		if err := ffjson.Unmarshal([]byte(jsonOrigins), &corsAllowedOrigins); err != nil {
			panic("invalid CORS_ALLOW_ORIGIN ENV, must be json encoded string array")
		}
	} else {
		corsAllowedOrigins = []string{"https://adminbeta.unwrappdapp.com", "https://admin.unwrappdapp.com"}
	}
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   corsAllowedOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch},
		Debug:            false,
	})
	router := mux.NewRouter().StrictSlash(true)
	router.Use(loggingMiddleware, corsMiddleware.Handler)
	router.Methods(http.MethodOptions).HandlerFunc(corsMiddleware.HandlerFunc)

	sr := router.PathPrefix("/v1").Subrouter()
	sr.HandleFunc("", RootHandler)
	sr.Use(auth.SessionMiddleware)

	ar := router.PathPrefix("/auth").Subrouter()
	auth.AddRoutes(ar)
	
	adr.Use(auth.SessionMiddleware)
	admin.AddRoutes(adr)

	return router
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	code int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.code = code
	lrw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userName := "-"

		app := appcontext.GetInstance()
		session, err := app.SESS_STORE.Get(r, "session")
		if err == nil {
			userSession := session.Values["user"]
			userString := fmt.Sprintf("%s", userSession)
			var user profile.Account
			ffjson.Unmarshal([]byte(userString), &user)
			userName = user.Email
		}

		start := time.Now()
		lrw := LoggingResponseWriter{w, 200}
		next.ServeHTTP(&lrw, r)
		elapsed := time.Since(start).Milliseconds()
		log.Printf("%s %s [%s] %s %s %d (%dms)", r.RemoteAddr, userName, start.Format(time.RFC3339), r.Method, r.URL.Path, lrw.code, elapsed)
	})
}

func main() {
	// Fetch DB config from environment
	dbhost := utils.GetEnvOrPanic("DB_HOST")
	dbname := utils.GetEnvOrPanic("POSTGRES_DB")
	dbuser := utils.GetEnvOrPanic("POSTGRES_USER")
	dbpassword := utils.GetEnvOrPanic("POSTGRES_PASSWORD")
	dbport, err := strconv.ParseInt(utils.GetEnvOrPanic("DB_PORT"), 10, 32)

	if err != nil {
		log.Fatal("Invalid DB Port")
	}

	// Initialize application context and set DB
	app := appcontext.GetInstance()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=America/New_York", dbhost, dbuser, dbpassword, dbname, dbport)
	retries := 10
	gormConfig := gorm.Config{}
	db, err := gorm.Open(postgres.Open(dsn), &gormConfig)
	// Retry a couple times on error in case the DB needs a chance to come up
	for ; err != nil; db, err = gorm.Open(postgres.Open(dsn), &gormConfig) {
		log.Printf("Error connecting to DB: %s (Retrying %d times)", err, retries)
		if retries <= 0 {
			log.Fatal("Out of retries")
		}

		retries--
		time.Sleep(5 * time.Second)
	}

	log.Println("Successfully connected to DB")
	app.DB = db

	db.SetupJoinTable(&wishlist.Wishlist{}, "Lines", &wishlist.WishlistLine{})

	// Run DB migrations
	migrations.Migrate()

	// store key for local dev can be anything/simple string, in prod/beta/k8s it is long/random
	session_store_key := utils.GetEnvOrPanic("SESSION_STORE_KEY")
	// instantiate interface which gormstore implements
	store := gormstore.NewOptions(
		db,
		gormstore.Options{
			TableName: "sessions",
		},
		[]byte(session_store_key))
	app.SESS_STORE = store

	// Load routes
	router := mainRouter()

	if err != nil {
		log.Fatalf("failed to set up the validator: %v", err)
	}

	// Start HTTP server
	log.Fatal(http.ListenAndServe(":8080", router))
}
