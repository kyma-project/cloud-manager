package util

import "math/rand"

func WithinProbability(probability int) bool {
	return rand.Intn(99)+1 <= probability
}
