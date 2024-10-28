package app

import (
	"math/rand/v2"
	"time"
)

// SleepMinPlusRandom sleeps for a duration that is randomly adjusted to be between the original duration and up to double that duration.
func SleepMinPlusRandom(minDuration time.Duration) {
	time.Sleep(time.Duration(float64(minDuration) * (1 + float64(rand.N(100))/100)))
}

// ReturnTrueXPercentOfTime returns true with a probability equal to the given percentage.
// It takes a float64 parameter 'percentage' which should be between 0 and 1.
// The function returns true approximately 'percentage' * 100% of the time.
//
// Example:
//
//	if ReturnTrueXPercentOfTime(0.25) {
//	    fmt.Println("This will print approximately 25% of the time")
//	}
//
// This function uses math/rand/v2, which does not require manual seeding.
func ReturnTrueXPercentOfTime(percentage float64) bool {
	return rand.Float64() < percentage
}
