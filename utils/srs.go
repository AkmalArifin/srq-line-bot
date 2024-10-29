package utils

import (
	"strings"
	"time"
)

func GetTimeReview(level int64) time.Time {
	now := time.Now().Truncate(time.Hour)

	switch level {
	case 1:
		return now
	case 2:
		return now.Add(4 * time.Hour)
	case 3:
		return now.Add(8 * time.Hour)
	case 4:
		return now.Add((24*1 - 1) * time.Hour)
	case 5:
		return now.Add((24*2 - 1) * time.Hour)
	case 6:
		return now.Add((24*7 - 1) * time.Hour)
	case 7:
		return now.Add((24*14 - 1) * time.Hour)
	case 8:
		return now.Add((24*30 - 1) * time.Hour)
	case 9:
		return now.Add((24*120 - 1) * time.Hour)
	case 10:
		return time.Time{}
	default:
		return now
	}
}

func GetNextLevel(level int64, ease string) int64 {
	var nextLevel int64
	switch strings.ToLower(ease) {
	case "easy": // easy
		nextLevel = level + 1
	case "good": // good
		nextLevel = level
	case "hard": //hard
		nextLevel = level - 1
	default:
		nextLevel = level
	}

	if nextLevel <= 0 {
		nextLevel = 1
	}
	return nextLevel
}
