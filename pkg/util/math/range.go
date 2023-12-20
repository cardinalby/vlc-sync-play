package mathutil

import "fmt"

type Range[T Numeric] struct {
	Min T
	Max T
}

func NewRangeUnordered[T Numeric](val1, val2 T) Range[T] {
	if val2 > val1 {
		return Range[T]{
			Min: val1,
			Max: val2,
		}
	}
	return Range[T]{
		Min: val2,
		Max: val1,
	}
}

func NewRangeMinWithLen[T Numeric](min T, length T) Range[T] {
	return Range[T]{
		Min: min,
		Max: min + length,
	}
}

func (r Range[T]) ToFloat64() Range[float64] {
	return Range[float64]{
		Min: float64(r.Min),
		Max: float64(r.Max),
	}
}

func (r Range[T]) Contains(value T) bool {
	return r.Min <= value && value <= r.Max
}

func (r Range[T]) Center() T {
	var denominator = T(2)
	return (r.Min + r.Max) / denominator
}

func (r Range[T]) Length() T {
	return r.Max - r.Min
}

func (r Range[T]) ContainsRange(other Range[T]) bool {
	return r.Min <= other.Min && other.Max <= r.Max
}

func (r Range[T]) Intersection(other Range[T]) (Range[T], bool) {
	if !r.HasIntersection(other) {
		return Range[T]{}, false
	}
	return Range[T]{
		Min: Max(r.Min, other.Min),
		Max: Min(r.Max, other.Max),
	}, true
}

func (r Range[T]) HasIntersection(other Range[T]) bool {
	return r.Min <= other.Max && other.Min <= r.Max
}

func (r Range[T]) IsValid() bool {
	return r.Min <= r.Max
}

func (r Range[T]) CapMax(cap T) Range[T] {
	return Range[T]{
		Min: r.Min,
		Max: Min(r.Max, cap),
	}
}

func (r Range[T]) CapLen(maxLength T) Range[T] {
	return Range[T]{
		Min: r.Min,
		Max: Min(r.Max, r.Min+maxLength),
	}
}

func (r Range[T]) Add(value T) Range[T] {
	return Range[T]{
		Min: r.Min + value,
		Max: r.Max + value,
	}
}

func (r Range[T]) AddRange(another Range[T]) Range[T] {
	return Range[T]{
		Min: r.Min + another.Min,
		Max: r.Max + another.Max,
	}
}

func (r Range[T]) Sub(value T) Range[T] {
	return Range[T]{
		Min: r.Min - value,
		Max: r.Max - value,
	}
}

func (r Range[T]) SubRange(another Range[T]) Range[T] {
	return Range[T]{
		Min: r.Min - another.Max,
		Max: r.Max - another.Min,
	}
}

func (r Range[T]) MultiplyF(value float64) Range[T] {
	return Range[T]{
		Min: T(float64(r.Min) * value),
		Max: T(float64(r.Max) * value),
	}
}

func (r Range[T]) MultiplyRangeF(another Range[float64]) Range[T] {
	return Range[T]{
		Min: T(float64(r.Min) * another.Min),
		Max: T(float64(r.Max) * another.Max),
	}
}

func (r Range[T]) DivF(value float64) Range[float64] {
	return Range[float64]{
		Min: float64(r.Min) / value,
		Max: float64(r.Max) / value,
	}
}

func (r Range[T]) String() string {
	return fmt.Sprintf("[%v, %v]", r.Min, r.Max)
}
