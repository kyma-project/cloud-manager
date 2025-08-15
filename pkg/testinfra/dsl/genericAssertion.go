package dsl

type GenericAssertion[T any] func(obj T) error

type GenericAssertions[T any] []GenericAssertion[T]

func (x GenericAssertions[T]) AssertObj(obj T) error {
	for _, f := range x {
		if err := f(obj); err != nil {
			return err
		}
	}
	return nil
}

func NewGenericAssertions[T any](items []GenericAssertion[T]) GenericAssertions[T] {
	return append(GenericAssertions[T]{}, items...)
}
