package util

import (
	"regexp"
	"time"
)

// MaxMinFloat returns the (max, min) of an array of ints.
func MaxMinInt(arr []int) (int, int) {
	if len(arr) == 0 {
		return -1, -1
	}
	max := arr[0]
	min := arr[0]
	for _, v := range arr {
		if v > max {
			max = v
		} else if v < min {
			min = v
		}
	}
	return max, min
}

// MaxMinFloat returns the (max, min) of an array of floats.
func MaxMinFloat(arr []float64) (float64, float64) {
	if len(arr) == 0 {
		return -1, -1
	}
	max := arr[0]
	min := arr[0]
	for _, v := range arr {
		if v > max {
			max = v
		} else if v < min {
			min = v
		}
	}
	return max, min
}

// Min returns the minimum time in an array of times.
func Min(arr []time.Time) (time.Time, int) {
	if len(arr) == 0 {
		return time.Time{}, -1
	}
	min := arr[0]
	minIdx := 0

	for i, v := range arr {
		if v.Before(min) {
			min = v
			minIdx = i
		}
	}
	return min, minIdx
}

// StripUsernames removes @ and #XXXX from usernames
// passed in, returning the results.
func StripUsernames(arr []string) []string {
	atRegex := regexp.MustCompile(`^@`)
	poundRegex := regexp.MustCompile(`#\d+$`)
	// Need to get rid of any user IDs and @s that they entered.
	for i := range arr {
		arr[i] = atRegex.ReplaceAllString(arr[i], "")
		arr[i] = poundRegex.ReplaceAllString(arr[i], "")
	}
	return arr
}

// Round rounds to a given unit place.
// 1 rounds to ones, 0.01 rounds to hundredths, 100 rounds to hundreds, etc.
func Round(x, unit float64) float64 {
	return float64(int64(x/unit+0.5)) * unit
}
