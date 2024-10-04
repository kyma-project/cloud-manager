package backupschedule

import (
	"math"
	"time"
)

var (
	ToleranceInterval = 1 * time.Second
)

func GetRemainingTime(to, from *time.Time) time.Duration {
	return GetRemainingTimeWithTolerance(to, from, ToleranceInterval)
}

func GetRemainingTimeWithTolerance(to, from *time.Time, tolerance time.Duration) time.Duration {
	if to == nil {
		to = &time.Time{}
	}
	if from == nil {
		from = &time.Time{}
	}

	timeLeft := to.Unix() - from.Unix()
	if math.Abs(float64(timeLeft)) <= tolerance.Seconds() {
		return 0
	}
	return time.Duration(timeLeft) * time.Second
}

func GetRemainingTimeFromNow(to *time.Time) time.Duration {
	now := time.Now().UTC()
	return GetRemainingTime(to, &now)
}

func GetRemainingTimeFromNowWithTolerance(to *time.Time, tolerance time.Duration) time.Duration {
	now := time.Now().UTC()
	return GetRemainingTimeWithTolerance(to, &now, tolerance)
}
