package client

import (
	"iter"
)

type Iterator[T any] interface {
	Next() (T, error)
	All() iter.Seq2[T, error]
}
