package chart

import (
	"fmt"

	"github.com/snyderks/chart"
)

// HorizontalLineSeries draws a horizontal line at the minimum value of the inner series.
type HorizontalLineSeries struct {
	Name        string
	Style       chart.Style
	YAxis       chart.YAxisType
	InnerSeries chart.ValuesProvider
	// Unused. This is to trick the renderer into thinking this is an annotation series.
	Annotations []chart.Value2

	Value float64
}

// GetName returns the name of the time series.
func (ms HorizontalLineSeries) GetName() string {
	return ms.Name
}

// GetStyle returns the line style.
func (ms HorizontalLineSeries) GetStyle() chart.Style {
	return ms.Style
}

// GetYAxis returns which YAxis the series draws on.
func (ms HorizontalLineSeries) GetYAxis() chart.YAxisType {
	return ms.YAxis
}

// Len returns the number of elements in the series.
func (ms HorizontalLineSeries) Len() int {
	return ms.InnerSeries.Len()
}

// GetValues gets a value at a given index.
func (ms *HorizontalLineSeries) GetValues(index int) (x, y float64) {
	x, _ = ms.InnerSeries.GetValues(index)
	y = ms.Value
	return
}

// Render renders the series.
func (ms *HorizontalLineSeries) Render(r chart.Renderer, canvasBox chart.Box, xrange, yrange chart.Range, defaults chart.Style) {
	style := ms.Style.InheritFrom(defaults)
	chart.Draw.LineSeries(r, canvasBox, xrange, yrange, style, ms)
}

// Validate validates the series.
func (ms *HorizontalLineSeries) Validate() error {
	if ms.InnerSeries == nil {
		return fmt.Errorf("Line series requires InnerSeries to be set")
	}
	return nil
}
