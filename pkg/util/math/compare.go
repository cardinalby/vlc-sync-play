package mathutil

type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

func Abs[T Numeric](value T) T {
	if value < 0 {
		return -value
	}
	return value
}

func Max[T Numeric](first T, other ...T) T {
	max := first

	for _, value := range other {
		if value > max {
			max = value
		}
	}

	return max
}

func Min[T Numeric](first T, other ...T) T {
	min := first

	for _, value := range other {
		if value < min {
			min = value
		}
	}

	return min
}

func NearestIndex[T Numeric](values []T, target T) int {
	if len(values) == 0 {
		return -1
	}
	minDiff := Abs(values[0] - target)
	minDiffIndex := 0

	for i := 1; i < len(values); i++ {
		if diff := Abs(values[i] - target); diff < minDiff {
			minDiff = diff
			minDiffIndex = i
		}
	}

	return minDiffIndex
}
