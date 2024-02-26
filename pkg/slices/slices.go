package slices

import (
	"slices"
)

func Count[V comparable](key V, values []V) uint {
	count := 0

	for _, v := range values {
		if key == v {
			count++
		}
	}

	return uint(count)
}

func Exist[V comparable](key V, values []V) bool {
	index := slices.Index(values, key)
	return index != -1
}
