package http

import (
	"net/http"

	"backend/internal/event"
	"backend/internal/logger"
	"backend/internal/repo"

	"gorm.io/gorm"
)

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewRouter(
	townRepo *repo.TownRepo,
	npcRepo *repo.NPCRepo,
	eventRepo *repo.EventRepo,
	locRepo *repo.LocationRepo,
	relRepo *repo.RelationshipRepo,
	db *gorm.DB,
	pub *event.Publisher,
	appLog *logger.AppLogger,
	directBroadcast func(eventType string, data map[string]any),
) *http.ServeMux {
	mux := http.NewServeMux()

	th := newTownHandler(townRepo)
	nh := newNPCHandler(npcRepo, townRepo, locRepo, relRepo)
	eh := newEventHandler(eventRepo, townRepo)
	ah := newAdminHandler(db)
	dh := newDemoHandler(pub)
	if directBroadcast != nil {
		dh.SetDirectBroadcast(directBroadcast)
	}
	// Build location name→ID map for demo direct broadcast
	if locs, err := locRepo.FindByTownID(1); err == nil {
		dh.locNameToID = make(map[string]uint)
		for _, l := range locs {
			dh.locNameToID[l.Name] = l.ID
		}
	}
	dih := newDiagHandler(db, GlobalDiag)

	mux.HandleFunc("/api/town/state", th.State)
	mux.HandleFunc("/api/npcs", nh.List)
	mux.HandleFunc("/api/npcs/", nh.Detail)
	mux.HandleFunc("/api/events", eh.List)
	mux.HandleFunc("/api/map", newMapHandler(locRepo))
	mux.HandleFunc("/api/admin/reset", ah.Reset)
	mux.HandleFunc("/api/admin/health", ah.Health)
	mux.HandleFunc("/api/demo/trigger", dh.Trigger)
	mux.HandleFunc("/api/diag/report", dih.Report)
	mux.HandleFunc("/api/diag/clear", dih.Clear)
	mux.HandleFunc("/api/diag/memory", dih.InspectMemory)
	mux.HandleFunc("/api/diag/gossip", dih.InspectGossip)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service":"CyberTown","version":"1.0"}`))
	})

	return mux
}
