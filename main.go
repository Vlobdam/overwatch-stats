package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/Vlobdam/overwatch-stats/dbHelper"
	"github.com/Vlobdam/overwatch-stats/handlers"
	"github.com/gorilla/mux"
)

var app *firebase.App
var credsJSON string
var url string
var port string

func init () {
	credsJSON = os.Getenv("FIREBASE_CREDENTIALS")
	url = os.Getenv("RTDB_URL")
	port = os.Getenv("PORT")
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusOK)
					return
			}

			h.ServeHTTP(w, r)
	}
}

func main() {
	ctx := context.Background()

	app = dbHelper.InitializeApp(ctx, credsJSON, url)

	// Create a new router
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/history", withCORS(handlers.GetStatsHandler(app, ctx, "matchHistory")))
	r.HandleFunc("/api/matchups", withCORS(handlers.GetStatsHandler(app, ctx, "matchUp")))
	r.HandleFunc("/api/synergy", withCORS(handlers.GetStatsHandler(app, ctx, "synergy")))
	r.HandleFunc("/api/maps", withCORS(handlers.GetStatsHandler(app, ctx, "mapPerformance")))
	r.HandleFunc("/api/match", withCORS(handlers.PostNewMatchHandler(app)))

	// Custom handler for React routes (SPA)
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve a static file, if it exists
		path := "./dist" + r.URL.Path
		if _, err := os.Stat(path); err == nil {
			http.ServeFile(w, r, path)
			return
		}

		// If the file does not exist, serve the SPA's index.html
		http.ServeFile(w, r, "./dist/index.html")
	})

	// Start the server
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
}