package emotion

func RecoverMood(currentMood string, hoursSinceChange float64) string {
	if currentMood == "content" {
		return "content"
	}
	if isPositiveMood(currentMood) && hoursSinceChange >= 1.0 {
		return "content"
	}
	if isNegativeMood(currentMood) && hoursSinceChange >= 2.0 {
		return "content"
	}
	return currentMood
}

func isPositiveMood(mood string) bool {
	switch mood {
	case "happy", "cheerful", "excited", "inspired", "jolly", "dreamy", "warm":
		return true
	}
	return false
}

func isNegativeMood(mood string) bool {
	switch mood {
	case "anxious", "worried", "sad", "angry", "tired":
		return true
	}
	return false
}

var ValidMoods = []string{
	"content", "happy", "cheerful", "excited", "inspired",
	"calm", "focused", "composed", "peaceful", "warm",
	"friendly", "confident", "caring", "dreamy", "steady",
	"jolly", "playful", "anxious", "worried", "sad", "angry", "tired",
	"curious", "cautious", "alert", "helpful",
}
