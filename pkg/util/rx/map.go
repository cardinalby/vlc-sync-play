package rx

type mappedObservable[T any, R any] struct {
	observable Observable[T]
	mapper     func(value T) R
}

func (o *mappedObservable[T, R]) GetValue() R {
	return o.mapper(o.observable.GetValue())
}

func (o *mappedObservable[T, R]) Subscribe(callback func(value R)) Subscription {
	return o.observable.Subscribe(func(value T) {
		callback(o.mapper(value))
	})
}

func Map[T any, R any](
	observable Observable[T],
	mapper func(value T) R,
) Observable[R] {
	return &mappedObservable[T, R]{
		observable: observable,
		mapper:     mapper,
	}
}
