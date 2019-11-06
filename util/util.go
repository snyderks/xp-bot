package util

import "regexp"

// MaxMin returns the (max, min) of an array of ints.
func MaxMin(arr []int) (int, int) {
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

// Strip usernames removes @ and #XXXX from usernames
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
