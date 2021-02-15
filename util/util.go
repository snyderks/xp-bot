package util

import (
	"regexp"
	"strings"
	"time"

	"github.com/snyderks/xp-bot/primitives"
)

// MaxMinInt returns the (max, min) of an array of ints.
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

// Avg returns the average delta between two XPs.
func Avg(hi, lo primitives.History) float64 {
	return float64(hi.XP-lo.XP) / (float64(hi.Date.Sub(lo.Date).Hours()) / 24.0)
}

// StripUsernames removes @ and #XXXX from usernames
// passed in, returning the results.
func StripUsernames(arr []string) []string {
	// Need to get rid of any chrome around the user ID.
	// Discord user IDs are of the form <@!################>
	// (amount of numbers should not be assumed)
	startRegex := regexp.MustCompile(`^<@!`)
	endRegex := regexp.MustCompile(`>$`)
	// Need to get rid of any user IDs and @s that they entered.
	for i := range arr {
		arr[i] = startRegex.ReplaceAllString(arr[i], "")
		arr[i] = endRegex.ReplaceAllString(arr[i], "")
	}
	return arr
}

// SpaceNormalizer guarantees that there will be no more than a single space
// between all arguments *within* the command, including usernames.
// For usernames, it also guarantees a space on either side of them.
func SpaceNormalizer(s string) string {
	// Discord user IDs are of the form <@!################>
	// (amount of numbers should not be assumed)
	startRegex := regexp.MustCompile(`\s*<@!`) // <@! with 0 or more spaces before it
	endRegex := regexp.MustCompile(`>\s*`)     // > with 0 or more spaces after it
	spaceRegex := regexp.MustCompile(`\s+`)    // 1 space or more
	beginningSpaceRegex := regexp.MustCompile(`^\s+`)
	endSpaceRegex := regexp.MustCompile(`\s+$`)

	// Order is going to matter here for efficiency.
	// 1. Add all spaces at the end of the usernames
	// 2. Add all spaces at the beginning of the usernames
	// 3. Remove all extra spaces between arguments
	// 4. Remove all spaces at the beginning and end
	s = startRegex.ReplaceAllLiteralString(s, " <@!")
	s = endRegex.ReplaceAllLiteralString(s, "> ")
	s = spaceRegex.ReplaceAllLiteralString(s, " ")
	s = beginningSpaceRegex.ReplaceAllLiteralString(s, "")
	s = endSpaceRegex.ReplaceAllLiteralString(s, "")
	return s
}

// Round rounds to a given unit place.
// 1 rounds to ones, 0.01 rounds to hundredths, 100 rounds to hundreds, etc.
func Round(x, unit float64) float64 {
	return float64(int64(x/unit+0.5)) * unit
}

// StringChecker checks whether a string matches any string in a slice.
// Set caseInsensitive to true to check strings with case-insensitivity;
// otherwise set it to false.
func StringChecker(s string, checkAgainst []string, caseInsensitive bool) bool {
	for _, el := range checkAgainst {
		if caseInsensitive {
			// EqualFold takes care of cases.
			if strings.EqualFold(s, el) {
				return true
			}
		} else {
			if s == el {
				return true
			}
		}
	}
	return false
}

// Between returns whether a given time is between start and end.
// If start and end are equal and/or end comes before start,
// it will always return false.
func Between(start, end, check *time.Time) bool {
	return check.After(*start) && check.Before(*end) ||
		check.Equal(*start) || check.Equal(*end)
}
