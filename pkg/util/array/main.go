package array

func Map[F any, T any](arr []F, transformFunc func(F) T) []T {
	res := make([]T, len(arr))
	for i, el := range arr {
		res[i] = transformFunc(el)
	}
	return res
}

func MapErr[F any, T any](arr []F, transformFunc func(F) (T, error)) ([]T, error) {
	var err error

	res := make([]T, len(arr))
	for i, el := range arr {
		res[i], err = transformFunc(el)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
