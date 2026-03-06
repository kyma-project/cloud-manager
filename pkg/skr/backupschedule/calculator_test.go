package backupschedule

import (
	"testing"
	"time"

	"github.com/gorhill/cronexpr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testclock "k8s.io/utils/clock/testing"
)

func TestScheduleCalculator_Now(t *testing.T) {
	fixed := time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC)
	fakeClock := testclock.NewFakeClock(fixed)
	calc := NewScheduleCalculator(fakeClock, 1*time.Second)

	assert.Equal(t, fixed, calc.Now())

	// After stepping the clock, Now() reflects the new time
	fakeClock.Step(5 * time.Minute)
	assert.Equal(t, fixed.Add(5*time.Minute), calc.Now())
}

func TestScheduleCalculator_GetRemainingTime(t *testing.T) {
	fixed := time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC)
	fakeClock := testclock.NewFakeClock(fixed)
	calc := NewScheduleCalculator(fakeClock, 5*time.Second)

	t.Run("target in future beyond tolerance", func(t *testing.T) {
		target := fixed.Add(10 * time.Minute)
		remaining := calc.GetRemainingTime(target)
		assert.Equal(t, 10*time.Minute, remaining)
	})

	t.Run("target in past beyond tolerance", func(t *testing.T) {
		target := fixed.Add(-10 * time.Minute)
		remaining := calc.GetRemainingTime(target)
		assert.Equal(t, -10*time.Minute, remaining)
	})

	t.Run("target within tolerance returns zero", func(t *testing.T) {
		target := fixed.Add(3 * time.Second)
		remaining := calc.GetRemainingTime(target)
		assert.Equal(t, time.Duration(0), remaining)
	})

	t.Run("target exactly at tolerance boundary returns zero", func(t *testing.T) {
		target := fixed.Add(5 * time.Second)
		remaining := calc.GetRemainingTime(target)
		assert.Equal(t, time.Duration(0), remaining)
	})
}

func TestScheduleCalculator_GetRemainingTimeWithTolerance(t *testing.T) {
	fixed := time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC)
	fakeClock := testclock.NewFakeClock(fixed)
	calc := NewScheduleCalculator(fakeClock, 1*time.Second)

	t.Run("custom tolerance", func(t *testing.T) {
		target := fixed.Add(30 * time.Second)
		remaining := calc.GetRemainingTimeWithTolerance(target, 1*time.Minute)
		assert.Equal(t, time.Duration(0), remaining)
	})

	t.Run("outside custom tolerance", func(t *testing.T) {
		target := fixed.Add(2 * time.Minute)
		remaining := calc.GetRemainingTimeWithTolerance(target, 1*time.Minute)
		assert.Equal(t, 2*time.Minute, remaining)
	})
}

func TestScheduleCalculator_NextRunTimes(t *testing.T) {
	fixed := time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC)
	fakeClock := testclock.NewFakeClock(fixed)
	calc := NewScheduleCalculator(fakeClock, 1*time.Second)

	// Every hour cron
	expr, err := cronexpr.Parse("0 * * * *")
	require.NoError(t, err)

	t.Run("nil start uses now", func(t *testing.T) {
		runs := calc.NextRunTimes(expr, nil, 3)
		require.Len(t, runs, 3)
		assert.Equal(t, time.Date(2026, 3, 5, 13, 0, 0, 0, time.UTC), runs[0])
		assert.Equal(t, time.Date(2026, 3, 5, 14, 0, 0, 0, time.UTC), runs[1])
		assert.Equal(t, time.Date(2026, 3, 5, 15, 0, 0, 0, time.UTC), runs[2])
	})

	t.Run("zero start uses now", func(t *testing.T) {
		zero := time.Time{}
		runs := calc.NextRunTimes(expr, &zero, 3)
		require.Len(t, runs, 3)
		assert.Equal(t, time.Date(2026, 3, 5, 13, 0, 0, 0, time.UTC), runs[0])
	})

	t.Run("past start uses now", func(t *testing.T) {
		past := fixed.Add(-1 * time.Hour)
		runs := calc.NextRunTimes(expr, &past, 2)
		require.Len(t, runs, 2)
		assert.Equal(t, time.Date(2026, 3, 5, 13, 0, 0, 0, time.UTC), runs[0])
	})

	t.Run("future start uses start", func(t *testing.T) {
		future := time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC)
		runs := calc.NextRunTimes(expr, &future, 2)
		require.Len(t, runs, 2)
		assert.Equal(t, time.Date(2026, 3, 6, 1, 0, 0, 0, time.UTC), runs[0])
		assert.Equal(t, time.Date(2026, 3, 6, 2, 0, 0, 0, time.UTC), runs[1])
	})

	t.Run("count of 1 returns single entry", func(t *testing.T) {
		runs := calc.NextRunTimes(expr, nil, 1)
		require.Len(t, runs, 1)
		assert.Equal(t, time.Date(2026, 3, 5, 13, 0, 0, 0, time.UTC), runs[0])
	})
}

func TestMaxSchedules(t *testing.T) {
	assert.Equal(t, uint(3), MaxSchedules)
}
