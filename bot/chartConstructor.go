package bot

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/snyderks/xp-bot/chart"
	"github.com/snyderks/xp-bot/db"
	"github.com/snyderks/xp-bot/logger"
	"github.com/snyderks/xp-bot/util"
)

// NeedMoreRecordsError is an error type designed to be returned to the user.
// Check for it and return it to the user if it appears.
var NeedMoreRecordsError = `The most recent record does not contain all of the 
top users. Please type t!top and try again.`

// RankLineChart retrieves records to construct a line chart of the top N users
// and returns the source and a list of usernames that couldn't be retrieved.
// Returns an error if the most recent day doesn't contain all the required
// records or if a generic error occurred.
func RankLineChart(c *db.DB, args *Args) (chart.LineChartSource, []string, error) {
	day, err := c.ReadNewestDay()
	if err != nil {
		return chart.LineChartSource{}, nil, errors.New(fmt.Sprint("failed to retrieve day", err.Error()))
	}
	// We need to have every single person in the range.
	if day.MinRank != 1 || day.MaxRank < args.Top {
		return chart.LineChartSource{}, nil, errors.New(NeedMoreRecordsError)
	}

	// Extract usernames
	people := make([]string, 0)
	for i, p := range day.People {
		if i == args.Top {
			break
		}
		people = append(people, p.UName)
	}

	// Get XP history for each person
	// Replace property on args with default if necessary
	if args.Days == 0 {
		args.Days = chart.GlobalChartConfig.DaysLimit
	}
	notFound, xpHistories, err := c.ReadPeople(people,
		args.Days)

	if err != nil {
		return chart.LineChartSource{}, nil,
			errors.New(fmt.Sprint("failed to retrieve people", err.Error()))
	}

	logger.Log.Info("Retrieved people. Failed to retrieve: ", notFound)

	series := make([][]float64, len(xpHistories))
	x := make([]time.Time, 0)
	seriesIdx := make([]int, len(xpHistories))

	// First, find the smallest date that starts all of the series.
	firstDates := make([]time.Time, len(xpHistories))
	for i, v := range xpHistories {
		firstDates[i] = v.History[0].Date
	}
	minDate, _ := util.Min(firstDates)

	// Track the max and min overall to pass off.
	overallMax := 0
	overallMin := math.MaxInt64

	// Now we step through all the histories to construct a set of series
	// and the X-axis.
	// We need to step through them as slowly as possible so that all
	// the XP is *aligned* and we have every single possible date for each
	// history. Any missing value will be filled by the previous value.
	for true {
		// Track how many of the series we're done with.
		offTheEnd := 0
		x = append(x, minDate)
		for i, x := range seriesIdx {
			if len(xpHistories[i].History) <= x {
				// Ran off the end. Append previous and don't advance.
				series[i] = append(series[i], float64(xpHistories[i].History[x-1].XP))
				offTheEnd++
			} else if xpHistories[i].History[x].Date.Equal(minDate) {
				// If equal, append and advance 1.
				series[i] = append(series[i], float64(xpHistories[i].History[x].XP))

				// Tracking the overall max and min...
				if xpHistories[i].History[x].XP > overallMax {
					overallMax = xpHistories[i].History[x].XP
				}
				if xpHistories[i].History[x].XP < overallMin {
					overallMin = xpHistories[i].History[x].XP
				}
				// Advance, advance!
				seriesIdx[i]++
			} else {
				// Haven't hit the next position yet. Append previous and don't advance.
				// Need to make sure we don't access the negative index just in case.
				zeroCase := int(math.Max(float64(x-1), 0))
				series[i] = append(series[i], float64(xpHistories[i].History[zeroCase].XP))
			}
		}

		if offTheEnd == len(xpHistories) {
			// All done constructing the series.
			break
		}

		// Get the next set of minimum dates and advance.
		for i, v := range xpHistories {
			idx := seriesIdx[i]
			if seriesIdx[i] >= len(xpHistories[i].History) {
				// What an interesting bug. If the very first person
				// runs off the end, it enters an infinite loop and
				// never stops. Therefore, if we hit the end of one
				// of them, set it to now (which will always be
				// higher than a date in the past).
				firstDates[i] = time.Now()
			} else {
				firstDates[i] = v.History[idx].Date
			}
		}
		minDate, _ = util.Min(firstDates)
	}

	return chart.LineChartSource{
			X:              x,
			Series:         series,
			Labels:         people,
			Title:          chart.GlobalChartConfig.RankChartTitle,
			LogScale:       true,
			ShowMilestones: true,
			Max:            float64(overallMax),
			Min:            float64(overallMin),
		},
		notFound, nil
}
