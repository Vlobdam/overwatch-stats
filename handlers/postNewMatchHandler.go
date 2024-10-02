package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	firebase "firebase.google.com/go/v4"
	db "firebase.google.com/go/v4/db"
	"github.com/vlobdam/overwatch-stats-backend/consts"
	"github.com/vlobdam/overwatch-stats-backend/dbHelper"
)

type MatchData struct {
	Map         string   `json:"map"`
	WinningTeam []string `json:"winning-team"`
	LosingTeam  []string `json:"losing-team"`
}

func PostNewMatchHandler(app *firebase.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		client := dbHelper.ConnectToRTDB(app, context.Background())

		var data MatchData
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		isValid := checkData(data)
		if !isValid {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		ref := client.NewRef("matchHistory")
		if _, err = ref.Push(context.Background(), data); err != nil {
			http.Error(w, "Failed to add match data", http.StatusInternalServerError)
			return
		}

		err = updateStats(data, client)
		if err != nil {
			http.Error(w, "Failed to update stats", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func checkData(data MatchData) bool {
	if !getIsValid([]string{data.Map}, consts.Maps) {
		return false
	}
	if !getIsValid(data.LosingTeam, consts.Heroes) {
		return false
	}
	if !getIsValid(data.WinningTeam, consts.Heroes) {
		return false
	}
	return true
}

func updateStats(match MatchData, client *db.Client) error {
	winningTeam := match.WinningTeam
	losingTeam := match.LosingTeam
	mapName := match.Map

	processMapPerformance(client, winningTeam, mapName, true)
	processMapPerformance(client, losingTeam, mapName, false)

	processMatchups(client, winningTeam, losingTeam)
	processSynergy(client, winningTeam, true)
	processSynergy(client, losingTeam, false)

	return nil
}

func processMapPerformance(client *db.Client, team []string, mapName string, isWinning bool) {
	for _, hero := range team {
		docID := fmt.Sprintf("mapPerformance/%s-on-%s", hero, mapName)
		ref := client.NewRef(docID)

		ref.Transaction(context.Background(), func(node db.TransactionNode) (interface{}, error) {
			var current map[string]int
			if err := node.Unmarshal(&current); err != nil || current == nil {
				current = map[string]int{
					"total": 0,
					"wins":  0,
				}
			}

			current["total"] += 1
			if isWinning {
				current["wins"] += 1
			}
			return current, nil
		})
	}
}

func processMatchups(client *db.Client, winningTeam []string, losingTeam []string) {
	for _, hero1 := range winningTeam {
		for _, hero2 := range losingTeam {
			winDocID := fmt.Sprintf("matchUp/%s-vs-%s", hero1, hero2)
			loseDocID := fmt.Sprintf("matchUp/%s-vs-%s", hero2, hero1)

			client.NewRef(winDocID).Transaction(context.Background(), func(node db.TransactionNode) (interface{}, error) {
				var current map[string]int
				if err := node.Unmarshal(&current); err != nil || current == nil {
					current = map[string]int{
						"total": 0,
						"wins":  0,
					}
				}
				current["total"] += 1
				current["wins"] += 1
				return current, nil
			})

			client.NewRef(loseDocID).Transaction(context.Background(), func(node db.TransactionNode) (interface{}, error) {
				var current map[string]int
				if err := node.Unmarshal(&current); err != nil || current == nil {
					current = map[string]int{
						"total": 0,
					}
				}
				current["total"] += 1
				return current, nil
			})
		}
	}
}

func processSynergy(client *db.Client, team []string, isWinning bool) {
	for i := 0; i < len(team)-1; i++ {
		for j := i + 1; j < len(team); j++ {
			hero1 := team[i]
			hero2 := team[j]

			docID1 := fmt.Sprintf("synergy/%s-with-%s", hero1, hero2)
			docID2 := fmt.Sprintf("synergy/%s-with-%s", hero2, hero1)

			client.NewRef(docID1).Transaction(context.Background(), func(node db.TransactionNode) (interface{}, error) {
				var current map[string]int
				if err := node.Unmarshal(&current); err != nil || current == nil {
					current = map[string]int{
						"total": 0,
						"wins":  0,
					}
				}
				current["total"] += 1
				if isWinning {
					current["wins"] += 1
				}
				return current, nil
			})

			client.NewRef(docID2).Transaction(context.Background(), func(node db.TransactionNode) (interface{}, error) {
				var current map[string]int
				if err := node.Unmarshal(&current); err != nil || current == nil {
					current = map[string]int{
						"total": 0,
						"wins":  0,
					}
				}
				current["total"] += 1
				if isWinning {
					current["wins"] += 1
				}
				return current, nil
			})
		}
	}
}

func getIsValid(values []string, validList []string) bool {
	for _, value := range values {
		isValueValid := false
		
		for _, v := range validList {
			if v == value {
				isValueValid = true
				break
			}
		}

		if !isValueValid {
			return false
		}
	}
	return true
}
