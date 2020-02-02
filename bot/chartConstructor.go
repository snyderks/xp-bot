package bot

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/snyderks/xp-bot/chart"
	"github.com/snyderks/xp-bot/db"
	"github.com/snyderks/xp-bot/logger"
	"github.com/snyderks/xp-bot/util"
	"github.com/snyderks/xp-bot/primitives"
)

// NeedMoreRecordsError is an error type designed to be returned to the user.
// Check for it and return it to the user if it appears.
var NeedMoreRecordsError = `The most recent record does not contain all of the 
top users. Please type t!top and try again.`

// FailedToRetrievePeopleError is an error type designed to be returned to the user.
// Check for it and return it to the user if it appears.
var FailedToRetrievePeopleError = `One or more usernames weren't correct. Try
fixing the following usernames:`

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

	series, x, overallMax, overallMin := constructSeries(xpHistories)

	if err != nil {
		return chart.LineChartSource{}, nil,
			errors.New(fmt.Sprint(FailedToRetrievePeopleError, strings.Join(notFound, ",")))
	}

	logger.Log.Info("Retrieved people. Failed to retrieve: ", notFound)

	return chart.LineChartSource{
			X:              x,
			Series:         series,
			Labels:         people,
			Title:          chart.GlobalChartConfig.RankChartTitle,
			LogScale:       true,
			ShowMilestones: true,
			Max:            float64(overallMax),
			Min:            float64(overallMin),
			King:           args.King,
		},
		notFound, nil
}

// UserLineChart retrieves records to construct a line chart of the users specified
// and returns the source and a list of usernames that couldn't be retrieved.
// Returns an error if a user wasn't found or if a generic error occurs.
func UserLineChart(c *db.DB, args *Args) (chart.LineChartSource, []string, error) {
	// Get XP history for each person
	// Replace property on args with default if necessary
	if args.Days == 0 {
		args.Days = chart.GlobalChartConfig.DaysLimit
	}
	notFound, xpHistories, err := c.ReadPeople(args.Usernames,
		args.Days)

	if err != nil {
		return chart.LineChartSource{}, nil,
			errors.New(fmt.Sprint(FailedToRetrievePeopleError, strings.Join(notFound, ",")))
	}

	series, x, overallMax, overallMin := constructSeries(xpHistories)

	logger.Log.Info("Retrieved people. Failed to retrieve: ", notFound)

	return chart.LineChartSource{
			X:      x,
			Series: series,
			Labels: args.Usernames,
			Title: fmt.Sprintf("Comparison of %s",
				strings.Join(args.Usernames, ", ")),
			LogScale:       false,
			ShowMilestones: true,
			Max:            float64(overallMax),
			Min:            float64(overallMin),
			King:           args.King,
		},
		notFound, nil
}

// SubLineChart retrieves records to construct a line chart with one user's
// XP subtracted from another user.
// Returns an error if a user wasn't found or if a generic error occurs.
func SubLineChart(c *db.DB, args *Args) (chart.LineChartSource, []string, error) {
	// Get XP history for each person
	// Replace property on args with default if necessary
	if args.Days == 0 {
		args.Days = chart.GlobalChartConfig.DaysLimit
	}
	notFound, xpHistories, err := c.ReadPeople(args.Usernames,
		args.Days)

	if err != nil {
		return chart.LineChartSource{}, nil,
			errors.New(fmt.Sprint(FailedToRetrievePeopleError, strings.Join(notFound, ",")))
	}

	series, x, _, _ := constructSeries(xpHistories)

	// No more than two users compared.
	if len(series) > 0 {
		series = series[0:2]
	}

	series, err = subtractor(series)

	logger.Log.Info("Retrieved people. Failed to retrieve: ", notFound)

	if notFound != nil && len(notFound) > 0 {
		return chart.LineChartSource{}, notFound, errors.New("failed to retrieve all users")
	}

	max, min := util.MaxMinFloat(series[0])

	fmt.Println(max, min)

	return chart.LineChartSource{
			X:      x,
			Series: series,
			Labels: []string{fmt.Sprintf("Comparison of %s",
				strings.Join(args.Usernames, ", "))},
			Title: fmt.Sprintf("Comparison of %s",
				strings.Join(args.Usernames, ", ")),
			LogScale:       false,
			ShowMilestones: true,
			Max:            max,
			Min:            min,
			King:           args.King,
		},
		notFound, nil
}

// subtractor takes two float series and subtracts the larger one's values
// from the smaller. The larger is determined by examining the last value
// in each series, with the greater value being the larger series.
func subtractor(series [][]float64) ([][]float64, error) {
	var large int
	var small int
	out := make([][]float64, 1)

	if len(series) != 2 {
		return out, errors.New("only two users are allowed")
	}

	if len(series[0]) != len(series[1]) {
		return out, errors.New("series must be the same length")
	}

	// Determine which is smaller and larger
	// Refer to them by their indices
	if series[0][len(series[0])-1] > series[1][len(series[1])-1] {
		large = 0
		small = 1
	} else {
		large = 1
		small = 0
	}

	out[0] = make([]float64, len(series[0]))

	for i := 0; i < len(series[0]); i++ {
		out[0][i] = series[large][i] - series[small][i]
	}

	return out, nil
}

// constructSeries creates a full set of points from a list of sparse arrays,
// replicating values in the sparse arrays by aligning the time values within them.
// returns the
func constructSeries(xpHistories []primitives.HistoryRange) (series [][]float64,
	x []time.Time, overallMin int, overallMax int) {
	series = make([][]float64, len(xpHistories))
	x = make([]time.Time, 0)
	seriesIdx := make([]int, len(xpHistories))

	// First, find the smallest date that starts all of the series.
	firstDates := make([]time.Time, len(xpHistories))
	for i, v := range xpHistories {
		firstDates[i] = v.History[0].Date
	}
	minDate, _ := util.Min(firstDates)

	// Track the max and min overall to pass off.
	overallMax = 0
	overallMin = math.MaxInt64

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
	return series, x, overallMax, overallMin
}
