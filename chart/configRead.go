package chart

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pelletier/go-toml"
)

// GlobalParams is the structure of the global parameters for every chart.
type GlobalParams struct {
	DaysLimit         int     `toml:"daysLimit"`
	RankChartTitle    string  `toml:"rankChartTitle"`
	BG                string  `toml:"bg"`
	FontColor         string  `toml:"fontColor"`
	AxesColor         string  `toml:"axesColor"`
	AxesFontSize      float64 `toml:"axesFontSize"`
	TitleFontSize     float64 `toml:"titleFontSize"`
	SeriesStrokeWidth float64 `toml:"seriesStrokeWidth"`
	Height            int     `toml:"height"`
	Width             int     `toml:"width"`
	DPI               float64 `toml:"dpi"`
}

type doc struct {
	ChartColors GlobalParams `toml:"chart"`
}

// Config is the object containing the configuration options for the chart.
var Config *toml.Tree

// GlobalChartConfig defines the set of parameters to be applied to every chart.
var GlobalChartConfig GlobalParams

var peopleColorsTable = "people-colors"

var colorsPath = os.Getenv("XP_BOT_COLORS_PATH")

func init() {
	b, err := ioutil.ReadFile(colorsPath)
	if err != nil {
		fmt.Println("Couldn't read colors config. Please make sure colors.toml",
			"exists in the same directory as the program.")
		panic(-1)
	}
	Config, err = toml.LoadBytes(b)
	if err != nil {
		fmt.Println("Couldn't parse colors config. Please make sure colors.toml",
			"is correctly formatted.")
		panic(-1)
	}

	// Take the document and put the global params
	// (shouldn't change) into an object
	d := doc{}
	toml.Unmarshal(b, &d)
	GlobalChartConfig = d.ChartColors
}

// GetUserColor returns the config color for a given user.
// Returns empty string and errors if the user isn't found.
func GetUserColor(t *toml.Tree, user string) (string, error) {
	color := t.Get(fmt.Sprintf("%s.%s", peopleColorsTable, user))
	if color == nil {
		return "", errors.New("color for that user not found")
	}
	return color.(string), nil
}
