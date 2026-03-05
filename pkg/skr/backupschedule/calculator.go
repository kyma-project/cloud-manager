package backupschedule

import (
	"math"
	"time"

	"github.com/gorhill/cronexpr"
	"k8s.io/utils/clock"
)

const MaxSchedules uint = 3

type ScheduleCalculator struct {
	Clock     clock.Clock
	Tolerance time.Duration
}

func NewScheduleCalculator(clk clock.Clock, tolerance time.Duration) *ScheduleCalculator {
	return &ScheduleCalculator{Clock: clk, Tolerance: tolerance}
}

func (c *ScheduleCalculator) Now() time.Time {
	return c.Clock.Now().UTC()
}

func (c *ScheduleCalculator) GetRemainingTime(target time.Time) time.Duration {
	return c.GetRemainingTimeWithTolerance(target, c.Tolerance)
}

func (c *ScheduleCalculator) GetRemainingTimeWithTolerance(target time.Time, tolerance time.Duration) time.Duration {
	now := c.Now()
	timeLeft := target.Unix() - now.Unix()
	if math.Abs(float64(timeLeft)) <= tolerance.Seconds() {
		return 0
	}
	return time.Duration(timeLeft) * time.Second
}

func (c *ScheduleCalculator) NextRunTimes(expr *cronexpr.Expression, start *time.Time, count uint) []time.Time {
	now := c.Now()
	if start != nil && !start.IsZero() && start.After(now) {
		return expr.NextN(start.UTC(), count)
	}
	return expr.NextN(now.UTC(), count)
}
