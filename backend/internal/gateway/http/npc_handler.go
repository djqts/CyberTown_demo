package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/repo"
)

type npcHandler struct {
	npcRepo *repo.NPCRepo
	townRepo *repo.TownRepo
	locRepo  *repo.LocationRepo
	relRepo  *repo.RelationshipRepo
}

func newNPCHandler(npcRepo *repo.NPCRepo, townRepo *repo.TownRepo, locRepo *repo.LocationRepo, relRepo *repo.RelationshipRepo) *npcHandler {
	return &npcHandler{npcRepo: npcRepo, townRepo: townRepo, locRepo: locRepo, relRepo: relRepo}
}

var npcColors = map[string]string{
	"埃德蒙": "#8B7355", "莉娜": "#78B860", "艾琳": "#C08880", "菲奥娜": "#F0C0D0",
	"奥托": "#686868", "克莱尔": "#F8F8F8", "杰克": "#C06050", "沃尔特": "#788860",
	"索菲亚": "#80B8D0", "皮埃尔": "#F0E890", "玛莎": "#E0A860", "卢卡斯": "#488880",
	"托马斯": "#8B7355", "米娅": "#F8E888", "薇拉": "#8868A0",
}

func (h *npcHandler) townID() uint {
	town, err := h.townRepo.FindFirst()
	if err != nil {
		return 1
	}
	return town.ID
}

func (h *npcHandler) List(w http.ResponseWriter, r *http.Request) {
	npcs, err := h.npcRepo.FindByTownID(h.townID())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to list npcs"})
		return
	}

	locNameMap := make(map[uint]string)
	if locs, err := h.locRepo.FindByTownID(int64(h.townID())); err == nil {
		for _, l := range locs {
			locNameMap[l.ID] = l.Name
		}
	}

	type npcSummary struct {
		ID            uint   `json:"id"`
		Name          string `json:"name"`
		Role          string `json:"role"`
		LocationID    uint   `json:"location_id"`
		LocationName  string `json:"location_name"`
		Mood          string `json:"mood"`
		Energy        int    `json:"energy"`
		CurrentGoal   string `json:"current_goal"`
		Gender        string `json:"gender"`
		AgeGroup      string `json:"age_group"`
		PortraitColor string `json:"portrait_color"`
	}

	summary := make([]npcSummary, 0, len(npcs))
	for _, npc := range npcs {
		color := npcColors[npc.Name]
		if color == "" {
			color = "#888888"
		}
		summary = append(summary, npcSummary{
			ID:            npc.ID,
			Name:          npc.Name,
			Role:          npc.Role,
			LocationID:    npc.LocationID,
			LocationName:  locNameMap[npc.LocationID],
			Mood:          npc.Mood,
			Energy:        npc.Energy,
			CurrentGoal:   npc.CurrentGoal,
			Gender:        npc.Gender,
			AgeGroup:      npc.AgeGroup,
			PortraitColor: color,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *npcHandler) Detail(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/npcs/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "npc id required"})
		return
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid npc id"})
		return
	}

	npc, err := h.npcRepo.FindByID(uint(id))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "npc not found"})
		return
	}

	color := npcColors[npc.Name]
	if color == "" {
		color = "#888888"
	}

	// Load relationships
	var relationships []map[string]any
	if rels, err := h.relRepo.FindByNPCID(npc.ID); err == nil {
		for _, rel := range rels {
			if target, err := h.npcRepo.FindByID(rel.TargetNPCID); err == nil {
				relationships = append(relationships, map[string]any{
					"npc_id":   target.ID,
					"npc_name": target.Name,
					"affinity": rel.Affinity,
					"trust":    rel.Trust,
					"tag":      rel.Tag,
				})
			}
		}
	}
	if relationships == nil {
		relationships = []map[string]any{}
	}

	resp := map[string]any{
		"id":             npc.ID,
		"name":           npc.Name,
		"role":           npc.Role,
		"personality":    npc.Personality,
		"catchphrase":    npc.Catchphrase,
		"appearance":     npc.Appearance,
		"gender":         npc.Gender,
		"age_group":      npc.AgeGroup,
		"mood":           npc.Mood,
		"energy":         npc.Energy,
		"current_goal":   npc.CurrentGoal,
		"status":         npc.Status,
		"location_id":    npc.LocationID,
		"portrait_color": color,
		"relationships":  relationships,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type eventLogHandler struct {
	eventRepo *repo.EventRepo
	townRepo  *repo.TownRepo
}

func newEventHandler(eventRepo *repo.EventRepo, townRepo *repo.TownRepo) *eventLogHandler {
	return &eventLogHandler{eventRepo: eventRepo, townRepo: townRepo}
}

func (h *eventLogHandler) townID() uint {
	town, err := h.townRepo.FindFirst()
	if err != nil {
		return 1
	}
	return town.ID
}

func (h *eventLogHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	events, err := h.eventRepo.FindRecent(h.townID(), limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to list events"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
