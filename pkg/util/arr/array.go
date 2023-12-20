package arr

func Map[T any, R any](collection []T, iteratee func(item T) R) []R {
	result := make([]R, len(collection))

	for i, item := range collection {
		result[i] = iteratee(item)
	}

	return result
}
