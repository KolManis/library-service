package repositories

// mapSlice применяет f к каждому элементу in, сохраняя порядок.
func mapSlice[T, U any](in []T, f func(T) U) []U {
	out := make([]U, 0, len(in))
	for _, v := range in {
		out = append(out, f(v))
	}
	return out
}
