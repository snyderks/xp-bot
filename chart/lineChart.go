package chart

import (
	"bytes"
	"math"
	"time"

	"github.com/snyderks/chart"
	"github.com/snyderks/chart/drawing"
	"github.com/snyderks/xp-bot/util"
)

// LineChartSource defines the data structure required to create a line chart.
type LineChartSource struct {
	// X is an array of dates to be displayed on the X-axis.
	X []time.Time
	// Series is an array of XP arrays, one array per person.
	// All series must be the same length.
	Series [][]float64
	// Labels is an array of usernames. Should be ordered
	// to correspond with Series and be the same length.
	Labels []string
	// Title is the title of the graph.
	Title string
	// LogScale determines whether the Y-axis of the graph
	// is printed on a log scale.
	LogScale bool
	// Max is the maximum XP value of Series.
	Max float64
	// Min is the minimum XP value of Series.
	Min float64
}

// Make returns a byte array encoded as an image containing the chart
// and a filename for the image.
func (src LineChartSource) Make() ([]byte, string) {
	graph := chart.Chart{
		DPI:    GlobalChartConfig.DPI,
		Width:  GlobalChartConfig.Width,
		Height: GlobalChartConfig.Height,
		Title:  src.Title,
		TitleStyle: chart.Style{
			Padding:   chart.NewBox(30, 0, 0, 0),
			FontColor: drawing.ColorFromHex(GlobalChartConfig.FontColor),
			FontSize:  GlobalChartConfig.TitleFontSize,
		},
		Background: chart.Style{
			Padding:   chart.NewBox(30, 200, 50, 50),
			FillColor: drawing.ColorFromHex(GlobalChartConfig.BG),
		},
		Canvas: chart.Style{
			FillColor: drawing.ColorFromHex(GlobalChartConfig.BG),
		},
		XAxis: chart.XAxis{
			Style: chart.Style{
				FontColor: drawing.ColorFromHex(GlobalChartConfig.FontColor),
				FontSize:  GlobalChartConfig.AxesFontSize,
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				FontColor: drawing.ColorFromHex(GlobalChartConfig.FontColor),
				FontSize:  GlobalChartConfig.AxesFontSize,
			},
		},
	}
	series := make([]chart.Series, 0)
	for i, person := range src.Labels {
		colorForPerson, err := GetUserColor(Config, person)
		if err != nil {
			colorForPerson = "ff0000"
		}
		series = append(series,
			chart.TimeSeries{
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex(colorForPerson),
					StrokeWidth: GlobalChartConfig.SeriesStrokeWidth,
					FontColor:   drawing.ColorFromHex(GlobalChartConfig.FontColor),
				},
				XValues: src.X,
				YValues: src.Series[i],
				Name:    person,
			})
	}

	// Give a 10% buffer at the top to the highest value
	// Round to the second-highest digit (i.e. 10000 rounds to nearest 1000)
	rangeMax := util.Round(src.Max*1.10, math.Pow10(int(math.Floor(math.Log10(src.Max))-1)))

	if src.LogScale {
		graph.YAxis.Range = &LogRange{Min: src.Min, Max: rangeMax}
	} else {
		graph.YAxis.Range = &chart.ContinuousRange{Min: src.Min, Max: rangeMax}
	}

	graph.Series = series

	graph.Elements = []chart.Renderable{
		chart.LegendLeft(&graph,
			chart.Style{
				FontSize:  GlobalChartConfig.AxesFontSize,
				FillColor: drawing.ColorFromHex(GlobalChartConfig.BG),
				FontColor: drawing.ColorFromHex(GlobalChartConfig.FontColor),
				Padding:   chart.NewBox(20, 20, 20, 20),
			}),
	}

	buf := bytes.NewBuffer([]byte{})
	graph.Render(chart.PNG, buf)
	return buf.Bytes(), "chart.png"
}
