package bot

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/snyderks/xp-bot/primitives"
)

// Parse takes in a string representation of the leaderboard
// and returns a list of the XP for each user.
func Parse(s string) (map[string]primitives.Person, error) {
	parseExp := regexp.MustCompile(`(\[)(\d+)(\].+)(#)(.+)([.\n\r\s]*)(Total Score: )(\d+)`)

	// Find all the results (-1 tells it never to stop)
	matches := parseExp.FindAllStringSubmatch(s, -1)

	if matches == nil {
		return nil, errors.New("no matches for the parser. Input was not of the correct format")
	}

	counts := make(map[string]primitives.Person, len(matches))

	for _, m := range matches {
		xp, err := strconv.Atoi(m[8])
		if err != nil {
			return nil, errors.New("an XP value could not be converted to a number")
		}
		rank, err := strconv.Atoi(m[2])
		if err != nil {
			return nil, errors.New("a rank could not be converted to a number")
		}

		counts[m[5]] = primitives.Person{UserID: m[5], XP: xp, Rank: rank}
	}

	return counts, nil
}
