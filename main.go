package main

import (
	"context"
	"log"
	"net/http"
	"os"

	firebase  "firebase.google.com/go/v4"
	"github.com/joho/godotenv"
	"github.com/Vlobdam/overwatch-stats/dbHelper"
	"github.com/Vlobdam/overwatch-stats/handlers"
)

var app *firebase.App
var credsJSON string
var url string

func init () {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
	credsJSON = os.Getenv("FIREBASE_CREDENTIALS")
	url = os.Getenv("RTDB_URL")
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

func main () {
	ctx := context.Background()
	
	app = dbHelper.InitializeApp(ctx, credsJSON, url)

	fs := http.FileServer(http.Dir("./dist"))
	http.Handle("/", fs)
	
	http.HandleFunc("/api/history", withCORS(handlers.GetStatsHandler(app, ctx, "matchHistory")))
	http.HandleFunc("/api/matchups", withCORS(handlers.GetStatsHandler(app, ctx, "matchUp")))
	http.HandleFunc("/api/synergy", withCORS(handlers.GetStatsHandler(app, ctx, "synergy")))
	http.HandleFunc("/api/maps", withCORS(handlers.GetStatsHandler(app, ctx, "mapPerformance")))

	http.HandleFunc("/api/match", withCORS(handlers.PostNewMatchHandler(app)))
	
	// Start HTTP server
	log.Fatal(http.ListenAndServe(":8080", nil))
}