package array

import (
	"slices"
	"strings"
)

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

func TrimmedUniqueStrings(input []string) []string {
	for i := range input {
		input[i] = strings.TrimSpace(input[i])
	}

	return UniqueValues(input)
}

func UniqueValues[T comparable](input []T) []T {
	uniqueSlice := []T{}
	for _, val := range input {
		if !slices.Contains(uniqueSlice, val) {
			uniqueSlice = append(uniqueSlice, val)
		}
	}

	return uniqueSlice
}
