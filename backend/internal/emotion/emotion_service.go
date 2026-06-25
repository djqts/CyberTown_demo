package emotion

import "backend/internal/logger"

type Service struct {
	appLog *logger.AppLogger
}

func NewService(appLog *logger.AppLogger) *Service {
	return &Service{appLog: appLog}
}

func (s *Service) ChangeMood(currentMood, newMood string) string {
	if newMood == "" || newMood == currentMood {
		return currentMood
	}
	if !isValidMood(newMood) {
		s.appLog.Warn("invalid mood ignored", "mood", newMood)
		return currentMood
	}
	return newMood
}

func isValidMood(mood string) bool {
	for _, m := range ValidMoods {
		if m == mood {
			return true
		}
	}
	return false
}
