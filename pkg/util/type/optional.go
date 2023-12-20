package typeutil

type Optional[T any] struct {
	Value    T
	HasValue bool
}

func NewOptional[T any](value T) Optional[T] {
	return Optional[T]{
		Value:    value,
		HasValue: true,
	}
}

func (o *Optional[T]) Set(value T) {
	o.Value = value
	o.HasValue = true
}

func (o *Optional[T]) Reset() {
	var empty T
	o.Value = empty
	o.HasValue = false
}

func (o *Optional[T]) Ptr() *T {
	if o.HasValue {
		return &o.Value
	}
	return nil
}
