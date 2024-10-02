package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"github.com/vlobdam/overwatch-stats-backend/dbHelper"
)

func GetStatsHandler(app *firebase.App, ctx context.Context, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client := dbHelper.ConnectToRTDB(app, ctx)

		ref := client.NewRef(key)

		var result map[string]interface{}
		if err := ref.Get(ctx, &result); err != nil {
			http.Error(w, "Failed to get data from database", http.StatusInternalServerError)
			log.Println("Error fetching data:", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}