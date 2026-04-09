package main

import (
	"math/rand"
	"time"
)

// ptr returns a pointer to the given string.
func ptr(s string) *string { return &s }

// ptrF returns a pointer to the given float64.
func ptrF(f float64) *float64 { return &f }

// ptrI returns a pointer to the given int.
func ptrI(i int) *int { return &i }

// daysAgo returns a time N days before now with a random hour offset.
func daysAgo(n int) time.Time {
	return time.Now().AddDate(0, 0, -n).Add(time.Duration(rand.Intn(24)) * time.Hour)
}

// daysAgoFixed returns a deterministic time N days before now at noon.
func daysAgoFixed(n int) time.Time {
	return time.Now().AddDate(0, 0, -n).Truncate(24*time.Hour).Add(12 * time.Hour)
}

// randDays returns a random time between 0 and maxDays ago.
func randDays(maxDays int) time.Time {
	return daysAgo(rand.Intn(maxDays) + 1)
}

// randChoice picks a random element from a slice.
func randChoice[T any](slice []T) T {
	return slice[rand.Intn(len(slice))]
}

// randChoiceN picks n random elements from a slice (without strict uniqueness).
func randChoiceN[T any](slice []T, n int) []T {
	result := make([]T, n)
	for i := range result {
		result[i] = slice[rand.Intn(len(slice))]
	}
	return result
}

// randInt returns a random int in [min, max].
func randInt(min, max int) int {
	if min >= max {
		return min
	}
	return min + rand.Intn(max-min+1)
}

// randFloat returns a random float64 in [min, max] rounded to nearest 1000.
func randCOP(minK, maxK int) float64 {
	return float64(randInt(minK, maxK) * 1000)
}
