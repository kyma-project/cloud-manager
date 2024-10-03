package backupschedule

import (
	"math"
	"time"
)

var (
	ToleranceInterval = 1 * time.Second
)

func GetRemainingTime(to, from *time.Time) time.Duration {
	if to == nil {
		to = &time.Time{}
	}
	if from == nil {
		from = &time.Time{}
	}

	timeLeft := to.Unix() - from.Unix()
	if math.Abs(float64(timeLeft)) < ToleranceInterval.Seconds() {
		return 0
	}
	return time.Duration(timeLeft) * time.Second
}

func GetRemainingTimeFromNow(to *time.Time) time.Duration {
	now := time.Now().UTC()
	return GetRemainingTime(to, &now)
}
