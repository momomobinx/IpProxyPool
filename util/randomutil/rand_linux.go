package randomutil

import (
	"math/rand"
	"time"
)

// RandInt64 get the random number in [min, max]
func RandInt64(min, max int64) int64 {
	if min >= max || max == 0 {
		return max
	}
	rng := rand.New(rand.NewSource(time.Now().Local().UnixNano()))
	num := rng.Int63n(max-min) + min
	return num
}

// RandInt get the random number in [min, max]
func RandInt(min, max int) int {
	if min >= max || max == 0 {
		return max
	}
	rng := rand.New(rand.NewSource(time.Now().Local().UnixNano()))
	num := rng.Intn(max-min) + min
	return num
}
