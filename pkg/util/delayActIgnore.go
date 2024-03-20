package util

import "cmp"

type DelayActIgnoreOutcome string

const (
	Unknown DelayActIgnoreOutcome = ""
	Delay   DelayActIgnoreOutcome = "delay"
	Act     DelayActIgnoreOutcome = "act"
	Ignore  DelayActIgnoreOutcome = "ignore"
	Error   DelayActIgnoreOutcome = "error"
)

type DelayActIgnore[T cmp.Ordered] interface {
	Case(v T) DelayActIgnoreOutcome
}

// DelayActIgnoreMap

func NewDelayActIgnoreMap[T cmp.Ordered](m map[T]DelayActIgnoreOutcome, notDefined DelayActIgnoreOutcome) DelayActIgnore[T] {
	return &delayActIgnoreMap[T]{
		notDefined: notDefined,
		m:          m,
	}
}

type delayActIgnoreMap[T cmp.Ordered] struct {
	notDefined DelayActIgnoreOutcome
	m          map[T]DelayActIgnoreOutcome
}

func (d *delayActIgnoreMap[T]) Case(v T) DelayActIgnoreOutcome {
	o, defined := d.m[v]
	if !defined {
		return d.notDefined
	}
	return o
}

// DelayActIgnoreMap Builder ========================

type DelayActIgnoreBuilder[T cmp.Ordered] struct {
	notDefined DelayActIgnoreOutcome
	m          map[T]DelayActIgnoreOutcome
}

func NewDelayActIgnoreBuilder[T cmp.Ordered](notDefined DelayActIgnoreOutcome) *DelayActIgnoreBuilder[T] {
	return &DelayActIgnoreBuilder[T]{
		notDefined: notDefined,
		m:          map[T]DelayActIgnoreOutcome{},
	}
}

func (b *DelayActIgnoreBuilder[T]) Ignore(states ...T) *DelayActIgnoreBuilder[T] {
	for _, s := range states {
		b.m[s] = Ignore
	}
	return b
}

func (b *DelayActIgnoreBuilder[T]) Act(states ...T) *DelayActIgnoreBuilder[T] {
	for _, s := range states {
		b.m[s] = Act
	}
	return b
}

func (b *DelayActIgnoreBuilder[T]) Delay(states ...T) *DelayActIgnoreBuilder[T] {
	for _, s := range states {
		b.m[s] = Delay
	}
	return b
}

func (b *DelayActIgnoreBuilder[T]) Error(states ...T) *DelayActIgnoreBuilder[T] {
	for _, s := range states {
		b.m[s] = Error
	}
	return b
}

func (b *DelayActIgnoreBuilder[T]) Build() DelayActIgnore[T] {
	return NewDelayActIgnoreMap(b.m, b.notDefined)
}
