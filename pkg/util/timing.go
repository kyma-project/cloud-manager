package util

import (
	"time"
)

/**
  real    tests
  100ms =   1ms
 1000ms =  10ms
10000ms = 100ms
60000ms = 600ms
*/

func SetRealTiming() {
	Timing = &timingStruct{divider: 1}
}

func SetSpeedyTimingForTests() {
	Timing = &timingStruct{divider: 100}
}

var Timing TimingIntf

func init() {
	SetRealTiming()
}

type TimingIntf interface {
	Divider() int64
	SetDivider(d int64)
	T100ms() time.Duration
	T1000ms() time.Duration
	T10000ms() time.Duration
	T60000ms() time.Duration
	T300000ms() time.Duration
}

func divideDuration(dur time.Duration, div int64) time.Duration {
	if div == 0 {
		return dur
	}
	res := time.Duration(int64(dur) / div)
	if res < time.Millisecond {
		return time.Millisecond
	}
	return res
}

type timingStruct struct {
	divider int64
}

func (t *timingStruct) Divider() int64 {
	return t.divider
}

func (t *timingStruct) SetDivider(d int64) {
	t.divider = d
}

func (t *timingStruct) T100ms() time.Duration {
	return divideDuration(100*time.Millisecond, t.divider)
}

func (t *timingStruct) T1000ms() time.Duration {
	return divideDuration(1000*time.Millisecond, t.divider)
}

func (t *timingStruct) T10000ms() time.Duration {
	return divideDuration(10000*time.Millisecond, t.divider)
}

func (t *timingStruct) T60000ms() time.Duration {
	return divideDuration(60000*time.Millisecond, t.divider)
}

func (t *timingStruct) T300000ms() time.Duration {
	return divideDuration(300000*time.Millisecond, t.divider)
}
