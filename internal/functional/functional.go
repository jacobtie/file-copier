package functional

func Filter[T any](s []T, predicate func(T) bool) []T {
	o := make([]T, 0)
	for _, e := range s {
		if predicate(e) {
			o = append(o, e)
		}
	}
	return o
}

func Map[I any, O any](s []I, mapper func(I) O) []O {
	o := make([]O, 0)
	for _, e := range s {
		o = append(o, mapper(e))
	}
	return o
}
