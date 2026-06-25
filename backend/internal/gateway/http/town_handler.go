package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"backend/internal/repo"
)

type townHandler struct {
	townRepo *repo.TownRepo
}

func newTownHandler(townRepo *repo.TownRepo) *townHandler {
	return &townHandler{townRepo: townRepo}
}

func (h *townHandler) State(w http.ResponseWriter, r *http.Request) {
	town, err := h.townRepo.FindFirst()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "town not found"})
		return
	}

	resp := map[string]any{
		"name":    town.Name,
		"day":     town.CurrentDay,
		"minute":  town.CurrentMinute,
		"time":    minuteToTimeStr(town.CurrentMinute),
		"season":  "春季",
		"weather": "晴",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func minuteToTimeStr(minute int) string {
	h := minute / 60
	m := minute % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}
