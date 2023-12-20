package mathutil

type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

func Max[T Numeric](lhs, rhs T) T {
	if lhs > rhs {
		return lhs
	}
	return rhs
}

func Min[T Numeric](lhs, rhs T) T {
	if lhs < rhs {
		return lhs
	}
	return rhs
}
