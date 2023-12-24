package rx

type Observable[T any] interface {
	GetValue() T
	Subscribe(callback func(value T)) Subscription
}

type Value[T any] interface {
	Observable[T]
	SetValue(value T)
}

type Subscription interface {
	Unsubscribe()
}
