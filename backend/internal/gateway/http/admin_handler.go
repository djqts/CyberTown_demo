package http

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

type adminHandler struct {
	db *gorm.DB
}

func newAdminHandler(db *gorm.DB) *adminHandler {
	return &adminHandler{db: db}
}

func (h *adminHandler) Reset(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tables := []string{
		"t_towns", "t_locations", "t_npcs", "t_npc_schedules",
		"t_chat_messages", "t_event_logs", "t_npc_relationships", "t_story_events",
	}
	for _, t := range tables {
		if err := h.db.Exec("TRUNCATE TABLE " + t + " CASCADE").Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "truncate failed: " + t})
			return
		}
	}
	// Reset sequences
	h.db.Exec("SELECT setval('t_towns_id_seq', 1, false)")
	h.db.Exec("SELECT setval('t_locations_id_seq', 37, true)")
	h.db.Exec("SELECT setval('t_npcs_id_seq', 33, true)")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "data reset, restart server to reseed"})
}

func (h *adminHandler) Health(w http.ResponseWriter, r *http.Request) {
	sqlDB, _ := h.db.DB()
	err := sqlDB.Ping()
	status := "healthy"
	if err != nil {
		status = "unhealthy: " + err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}
