package bot

import (
	"errors"
	"time"

	"github.com/snyderks/xp-bot/chart"
	"github.com/snyderks/xp-bot/util"
	"github.com/snyderks/xp-bot/primitives"
)

// TimeRange defines a start and end interval in months and days.
type TimeRange struct {
	MonthsAgoStart int
	DaysAgoStart   int
	MonthsAgoEnd   int
	DaysAgoEnd     int
}

// AverageForTimeRange gives an average daily xp for an interval of time.
// Returns an error if there isn't enough time to determine a value.
func AverageForTimeRange(interval TimeRange,
	history primitives.HistoryRange) (float64, error) {
	now := time.Now()
	start := now.AddDate(0, -interval.MonthsAgoStart, -interval.DaysAgoStart)
	end := now.AddDate(0, -interval.MonthsAgoEnd, -interval.DaysAgoEnd)

	subrange := make([]primitives.History, 0)
	for _, v := range history.History {
		if util.Between(&start, &end, &v.Date) {
			subrange = append(subrange, v)
		}
	}

	if len(subrange) == 0 {
		return -1,
			errors.New("nothing in this time range. Can't determine a mean")
	}

	return util.Avg(subrange[0], subrange[len(subrange)-1]), nil
}

// ExpectedDate returns the estimated date of reaching a given XP threshold
// from the current date. Returns an error if the XP values are equal, if
// toXP is smaller than fromCurrentXP, or if dailyXP is less than 1.
func ExpectedDate(fromCurrentXP, toXP, dailyXP int) (time.Time, error) {
	amt := toXP - fromCurrentXP
	if amt <= 0 {
		return time.Time{}, errors.New("Can't return for an amount leq zero")
	}
	if dailyXP <= 0 {
		return time.Time{},
			errors.New("Can't get expected date without gaining XP")
	}

	days := amt / dailyXP

	return time.Now().AddDate(0, 0, days), nil
}

// NextXPTier returns the next XP milestone based on current XP.
// Returns an error if they have no XP tier to reach.
func NextXPTier(currentXP int, XPList []chart.Milestone) (chart.Milestone, error) {
	for _, m := range XPList {
		if m.XP > currentXP {
			return m, nil
		}
	}
	return chart.Milestone{}, errors.New("No milestones left to reach")
}
