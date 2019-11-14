package chart

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/snyderks/chart"
	"github.com/snyderks/chart/drawing"
	"github.com/snyderks/xp-bot/util"
)

var fontPath = "./Ubuntu-Merged.ttf"
var font *truetype.Font

func init() {
	f, err := GetFont(fontPath)
	if err != nil {
		panic(err)
	}
	font = f
}

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
	// ShowMilestones tells the line chart whether to show the milestone lines.
	ShowMilestones bool
}

// Make returns a byte array encoded as an image containing the chart
// and a filename for the image.
func (src LineChartSource) Make() ([]byte, string) {
	// Give a 10% buffer at the top to the highest value
	// Round to the second-highest digit (i.e. 10000 rounds to nearest 1000)
	rangeMax := util.Round(src.Max*1.10, math.Pow10(int(math.Floor(math.Log10(src.Max))-1)))

	src.Max = rangeMax

	graph := chart.Chart{
		DPI:    GlobalChartConfig.DPI,
		Width:  GlobalChartConfig.Width,
		Height: GlobalChartConfig.Height,
		Title:  src.Title,
		Font:   font,
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
			Ticks: getYAxisTicks(src.Min, src.Max, src.LogScale),
			Style: chart.Style{
				FontColor: drawing.ColorFromHex(GlobalChartConfig.FontColor),
				FontSize:  GlobalChartConfig.AxesFontSize,
			},
		},
	}
	series := make([]chart.Series, 0)
	colorIdx := 0
	var firstSeries chart.TimeSeries
	for i, person := range src.Labels {
		var colorForPerson drawing.Color
		result, err := GetUserColor(Config, person)
		if err != nil {
			colorForPerson = chart.ColorPalette.GetSeriesColor(chart.DefaultColorPalette, colorIdx)
			colorIdx++
		} else {
			colorForPerson = drawing.ColorFromHex(result)
		}
		s := chart.TimeSeries{
			Style: chart.Style{
				StrokeColor: colorForPerson,
				StrokeWidth: GlobalChartConfig.SeriesStrokeWidth,
				FontColor:   drawing.ColorFromHex(GlobalChartConfig.FontColor),
			},
			XValues: src.X,
			YValues: src.Series[i],
			Name:    person,
		}
		series = append(series, s)
		if i == 0 {
			firstSeries = s
		}
	}

	if src.LogScale {
		graph.YAxis.Range = &LogRange{Min: src.Min, Max: rangeMax}
	} else {
		graph.YAxis.Range = &chart.ContinuousRange{Min: src.Min, Max: rangeMax}
	}

	if src.ShowMilestones {
		for _, ms := range GlobalChartConfig.Milestones {
			if float64(ms.XP) < src.Max && float64(ms.XP) > src.Min {
				line := &HorizontalLineSeries{
					Style: chart.Style{
						HiddenOnLegend:  true,
						StrokeColor:     chart.ColorLightGray,
						StrokeDashArray: []float64{5.0, 5.0},
						FontSize:        GlobalChartConfig.AxesFontSize,
					},
					Name:        ms.Name,
					InnerSeries: firstSeries,
					Value:       float64(ms.XP)}
				series = append(series, line)
				series = append(series, chart.LastValueAnnotationSeries(line))
			}
		}
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

func getYAxisTicks(min float64, max float64, log bool) []chart.Tick {
	var interval float64
	if log {
		interval = (math.Log10(max) - math.Log10(min)) / float64(GlobalChartConfig.TickNum)
	} else {
		interval = (max - min) / float64(GlobalChartConfig.TickNum)
	}

	ticks := make([]chart.Tick, GlobalChartConfig.TickNum+1)
	for i := 0; i <= GlobalChartConfig.TickNum; i++ {
		var next float64
		if log {
			next = math.Pow(10, math.Log10(min)+interval*float64(i))
		} else {
			next = min + interval*float64(i)
		}

		next = util.Round(next, math.Pow10(int(math.Floor(math.Log10(next))-1)))

		ticks[i] = chart.Tick{Value: next, Label: strconv.Itoa(int(next))}
	}
	fmt.Println(ticks)
	return ticks
}
