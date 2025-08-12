package value

func Coalesce[T any](from *T, to T) T {
	if from == nil {
		return to
	}

	return to
}
