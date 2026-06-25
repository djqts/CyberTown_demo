package story

import (
	"encoding/json"
	"math/rand"
	"time"
)

type TriggerCondition struct {
	TimeRange     string  `json:"time_range"`
	Probability   float64 `json:"probability"`
	CooldownHours int     `json:"cooldown_hours"`
	TriggerDay    string  `json:"trigger_day,omitempty"`
}

func ShouldTrigger(condJSON string, minuteOfDay int, lastTriggeredAt *time.Time) bool {
	var cond TriggerCondition
	if err := json.Unmarshal([]byte(condJSON), &cond); err != nil {
		return false
	}

	if cond.CooldownHours > 0 && lastTriggeredAt != nil {
		if time.Since(*lastTriggeredAt).Hours() < float64(cond.CooldownHours) {
			return false
		}
	}

	if !matchTimeRange(cond.TimeRange, minuteOfDay) {
		return false
	}

	if cond.Probability < 1.0 {
		return rand.Float64() < cond.Probability
	}

	return true
}

func matchTimeRange(timeRange string, minuteOfDay int) bool {
	switch timeRange {
	case "morning":
		return minuteOfDay >= 360 && minuteOfDay < 720
	case "noon":
		return minuteOfDay >= 660 && minuteOfDay < 840
	case "afternoon":
		return minuteOfDay >= 720 && minuteOfDay < 1080
	case "evening":
		return minuteOfDay >= 1080 && minuteOfDay < 1320
	case "any":
		return true
	}
	return true
}
