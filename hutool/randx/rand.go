package randx

import (
	"math/rand"
)

func SelectRandCount[T any](arr []T, count int) []T {
	if count > len(arr) {
		return arr
	}
	perm := rand.Perm(len(arr))
	result := make([]T, 0, count)
	for i := 0; i < count; i++ {
		result = append(result, arr[perm[i]])
	}
	return result
}
