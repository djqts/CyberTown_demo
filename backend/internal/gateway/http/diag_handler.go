package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// WSMsgTrace records a WebSocket message that was broadcast.
type WSMsgTrace struct {
	Time      time.Time `json:"time"`
	Type      string    `json:"type"`
	NpcID     uint      `json:"npc_id,omitempty"`
	NpcName   string    `json:"npc_name,omitempty"`
	FromLoc   string    `json:"from_location,omitempty"`
	ToLoc     string    `json:"to_location,omitempty"`
	FromLocID uint      `json:"from_location_id,omitempty"`
	ToLocID   uint      `json:"to_location_id,omitempty"`
}

// DiagCollector collects diagnostic data across layers.
type DiagCollector struct {
	mu          sync.Mutex
	WSTraces    []WSMsgTrace `json:"ws_traces"`
	PublishCnt  int          `json:"publish_count"`
	ErrorCnt    int          `json:"error_count"`
	StartTime   time.Time    `json:"start_time"`
}

var GlobalDiag = &DiagCollector{StartTime: time.Now()}

// diagRedis is set by main.go to enable memory inspection.
var diagRedis *redis.Client

// SetDiagRedis wires the Redis client for /api/diag/memory endpoint.
func SetDiagRedis(rc *redis.Client) { diagRedis = rc }

// diagGossipFn is set by main.go to enable gossip inspection.
var diagGossipFn func(npcID uint) string

// SetDiagGossip wires the social gossip source for /api/diag/gossip endpoint.
func SetDiagGossip(fn func(npcID uint) string) { diagGossipFn = fn }

func (d *DiagCollector) RecordBroadcast(msgType string, data map[string]any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	t := WSMsgTrace{Time: time.Now(), Type: msgType}
	if v, ok := data["npc_id"].(uint); ok {
		t.NpcID = v
	}
	if v, ok := data["npc_name"].(string); ok {
		t.NpcName = v
	}
	if v, ok := data["from_location"].(string); ok {
		t.FromLoc = v
	}
	if v, ok := data["to_location"].(string); ok {
		t.ToLoc = v
	}
	if v, ok := data["from_location_id"].(uint); ok {
		t.FromLocID = v
	}
	if v, ok := data["to_location_id"].(uint); ok {
		t.ToLocID = v
	}
	d.WSTraces = append(d.WSTraces, t)
	if len(d.WSTraces) > 200 {
		d.WSTraces = d.WSTraces[len(d.WSTraces)-200:]
	}
	d.PublishCnt++
}

func (d *DiagCollector) RecordError() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.ErrorCnt++
}

type diagHandler struct {
	db        *gorm.DB
	collector *DiagCollector
}

func newDiagHandler(db *gorm.DB, collector *DiagCollector) *diagHandler {
	return &diagHandler{db: db, collector: collector}
}

func (h *diagHandler) Report(w http.ResponseWriter, r *http.Request) {
	h.collector.mu.Lock()
	// Deep-copy traces to avoid data race during JSON serialization
	tracesCopy := make([]WSMsgTrace, len(h.collector.WSTraces))
	copy(tracesCopy, h.collector.WSTraces)
	report := map[string]any{
		"uptime_sec":   time.Since(h.collector.StartTime).Seconds(),
		"publish_cnt":  h.collector.PublishCnt,
		"error_cnt":    h.collector.ErrorCnt,
		"ws_trace_cnt": len(tracesCopy),
		"ws_traces":    tracesCopy,
	}
	h.collector.mu.Unlock()

	// DB stats
	var eventCount int64
	h.db.Model(&struct{ ID uint }{})
	// count event_logs
	var ec struct{ Count int64 }
	h.db.Raw("SELECT COUNT(*) as count FROM t_event_logs").Scan(&ec)
	eventCount = ec.Count

	var npcStates []map[string]any
	rows, _ := h.db.Raw("SELECT id, name, location_id, mood, current_goal, status FROM t_npcs WHERE town_id = 1 ORDER BY id").Rows()
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id, locID uint
			var name, mood, goal, status string
			rows.Scan(&id, &name, &locID, &mood, &goal, &status)
			npcStates = append(npcStates, map[string]any{
				"id": id, "name": name, "location_id": locID,
				"mood": mood, "current_goal": goal, "status": status,
			})
		}
	}

	report["event_log_count"] = eventCount
	report["npc_states"] = npcStates

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *diagHandler) Clear(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	h.collector.mu.Lock()
	h.collector.WSTraces = nil
	h.collector.PublishCnt = 0
	h.collector.ErrorCnt = 0
	h.collector.StartTime = time.Now()
	h.collector.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}

// InspectMemory shows short-term memory from Redis.
// GET /api/diag/memory?npc_id=35&user_token=web-user
func (h *diagHandler) InspectMemory(w http.ResponseWriter, r *http.Request) {
	if diagRedis == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "redis not configured"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	npcID := r.URL.Query().Get("npc_id")
	userToken := r.URL.Query().Get("user_token")

	var keys []string
	if npcID != "" && userToken != "" {
		keys = []string{"chat:short:" + npcID + ":" + userToken}
	} else if npcID != "" {
		var err error
		keys, err = diagRedis.Keys(ctx, "chat:short:"+npcID+":*").Result()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
	} else {
		var err error
		keys, err = diagRedis.Keys(ctx, "chat:short:*").Result()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		if len(keys) > 50 {
			keys = keys[:50]
		}
	}

	result := make(map[string][]map[string]any)
	for _, key := range keys {
		vals, err := diagRedis.LRange(ctx, key, 0, 10).Result()
		if err != nil {
			continue
		}
		var msgs []map[string]any
		for _, v := range vals {
			var m map[string]any
			json.Unmarshal([]byte(v), &m)
			msgs = append(msgs, m)
		}
		for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
			msgs[i], msgs[j] = msgs[j], msgs[i]
		}
		result[key] = msgs
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"key_count": len(keys),
		"memories":  result,
	})
}

// InspectGossip returns social gossip for an NPC.
// GET /api/diag/gossip?npc_id=35
func (h *diagHandler) InspectGossip(w http.ResponseWriter, r *http.Request) {
	if diagGossipFn == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"gossip": "", "note": "gossip source not wired"})
		return
	}
	npcID, _ := strconv.Atoi(r.URL.Query().Get("npc_id"))
	gossip := diagGossipFn(uint(npcID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"npc_id": npcID,
		"gossip": gossip,
	})
}
