package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vitwit/authz-apps/voting-bot/database"
)

func GetRewardsHandler(db *database.Sqlitedb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		params := r.URL.Query()
		chainId := params.Get("id")
		date := params.Get("date")

		rewards, err := db.GetRewards(chainId, date)
		if err != nil {
			http.Error(w, fmt.Errorf("error while getting rewards: %w", err).Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(rewards)
		if err != nil {
			http.Error(w, fmt.Errorf("error while encoding rewards: %w", err).Error(), http.StatusInternalServerError)
			return
		}
	}
}

func RetrieveProposalsHandler(db *database.Sqlitedb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		params := r.URL.Query()
		chainName := mux.Vars(r)["chainName"]
		start := params.Get("start")
		end := params.Get("end")

		proposals, err := db.GetProposals(chainName, start, end)
		if err != nil {
			http.Error(w, fmt.Errorf("error while getting proposals: %w", err).Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(proposals)
		if err != nil {
			http.Error(w, fmt.Errorf("error while encoding proposals: %w", err).Error(), http.StatusInternalServerError)
			return
		}
	}
}

func RetrieveProposalsForAllNetworksHandler(db *database.Sqlitedb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		params := r.URL.Query()
		start := params.Get("start")
		end := params.Get("end")

		proposals, err := db.GetProposalsForAllNetworks(start, end)
		if err != nil {
			http.Error(w, fmt.Errorf("error while getting proposals for all networks: %w", err).Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(proposals)
		if err != nil {
			http.Error(w, fmt.Errorf("error while encoding proposals for all networks: %w", err).Error(), http.StatusInternalServerError)
			return
		}
	}
}
