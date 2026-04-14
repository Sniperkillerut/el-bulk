package main

import (
	"math/rand"
	"time"
)

// ptr returns a pointer to the given string.
func ptr(s string) *string { return &s }

// daysAgo returns a time N days before now with a random hour offset.
func daysAgo(n int) time.Time {
	return time.Now().AddDate(0, 0, -n).Add(time.Duration(rand.Intn(24)) * time.Hour)
}

// daysAgoFixed returns a deterministic time N days before now at noon.
func daysAgoFixed(n int) time.Time {
	return time.Now().AddDate(0, 0, -n).Truncate(24*time.Hour).Add(12 * time.Hour)
}

// randInt returns a random int in [min, max].
func randInt(min, max int) int {
	if min >= max {
		return min
	}
	return min + rand.Intn(max-min+1)
}
