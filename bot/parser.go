package bot

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/snyderks/xp-bot/db"
)

// Parse takes in a string representation of the leaderboard
// and returns a list of the XP for each user.
func Parse(s string) (map[string]db.Person, error) {
	parseExp := regexp.MustCompile(`(\[)(\d+)(\].+)(#)(.+)([.\n\r\s]*)(Total Score: )(\d+)`)

	// Find all the results (-1 tells it never to stop)
	matches := parseExp.FindAllStringSubmatch(s, -1)

	if matches == nil {
		return nil, errors.New("no matches for the parser. Input was not of the correct format")
	}
	print(matches)

	counts := make(map[string]db.Person, len(matches))

	for _, m := range matches {
		xp, err := strconv.Atoi(m[8])
		if err != nil {
			return nil, errors.New("an XP value could not be converted to a number")
		}
		rank, err := strconv.Atoi(m[2])
		if err != nil {
			return nil, errors.New("a rank could not be converted to a number")
		}

		counts[m[5]] = db.Person{UName: m[5], XP: xp, Rank: rank}
	}

	return counts, nil
}
