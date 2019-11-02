package bot

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

// Person is the XP for a given uname.
type Person struct {
	UName string
	XP    int
	Rank  int
}

// History is a record of a user's XP at a given moment.
type History struct {
	XP   int
	Date time.Time
}

// HistoryRange is a list of moments of a user's XP.
type HistoryRange struct {
	History []History
	UName   string
}

// Parse takes in a string representation of the leaderboard
// and returns a list of the XP for each user.
func Parse(s string) ([]Person, error) {
	parseExp := regexp.MustCompile(`(#)(.+)([.\n\r\s]*)(Total Score: )(\d+)`)

	// Find all the results (-1 tells it never to stop)
	matches := parseExp.FindAllStringSubmatch(s, -1)

	if matches == nil {
		return nil, errors.New("no matches for the parser. Input was not of the correct format")
	}
	print(matches)

	counts := make([]Person, len(matches))

	for i, m := range matches {
		xp, err := strconv.Atoi(m[5])
		if err != nil {
			return nil, errors.New("an XP value could not be converted to a number")
		}
		counts[i] = Person{m[2], xp, 1} // TODO: add rank to the output
	}

	return counts, nil
}
