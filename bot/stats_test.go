package bot

import (
	"math"
	"testing"
	"time"

	"github.com/snyderks/xp-bot/primitives"
)

var now = time.Now()
var history = primitives.HistoryRange{
	History: []primitives.History{
		primitives.History{XP: 1, Date: now.AddDate(0, -1, 0)},
		primitives.History{XP: 20, Date: now.AddDate(0, 0, -15)},
		primitives.History{XP: 25, Date: now.AddDate(0, 0, -14).Add(time.Minute * 20)},
		primitives.History{XP: 48, Date: now.AddDate(0, 0, -7).Add(time.Minute * 10)},
		primitives.History{XP: 50, Date: now.AddDate(0, 0, -6)},
		primitives.History{XP: 100, Date: now},
	},
}

func TestAverageForTimeRange(t *testing.T) {
	t.Run("this week", func(t *testing.T) {
		interval := TimeRange{MonthsAgoStart: 0, MonthsAgoEnd: 0,
			DaysAgoStart: 7, DaysAgoEnd: 0}
		result, err := AverageForTimeRange(interval, history)

		if err != nil {
			t.Error(err)
			return
		}

		// (100 - 48) / 7
		correct := 7.4285

		// Almost equal
		if math.Abs(correct-result) > 0.01 {
			t.Errorf("Got %f for 7 days xp. Expected %f.", result, correct)
		}
	})

	t.Run("last week", func(t *testing.T) {
		interval := TimeRange{MonthsAgoStart: 0, MonthsAgoEnd: 0,
			DaysAgoStart: 14, DaysAgoEnd: 7}
		result, err := AverageForTimeRange(interval, history)

		if err != nil {
			t.Error(err)
			return
		}

		// (48 - 25) / 7
		correct := 3.286

		// Almost equal
		if math.Abs(correct-result) > 0.01 {
			t.Errorf("Got %f for 7 days xp. Expected %f.", result, correct)
		}
	})
}
