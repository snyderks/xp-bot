package chart

import (
	"fmt"

	"github.com/snyderks/chart"
)

// LastValueLabeledAnnotationSeries returns an annotation series of just the last value of a value provider.
// label will be prepended to the value as: label – value
func LastValueLabeledAnnotationSeries(innerSeries chart.ValuesProvider,
	label string, vfs ...chart.ValueFormatter) chart.AnnotationSeries {

	var lastValue chart.Value2
	if typed, isTyped := innerSeries.(chart.LastValuesProvider); isTyped {
		lastValue.XValue, lastValue.YValue = typed.GetLastValues()
		lastValue.Label = fmt.Sprintf("%s – %d", label, int(lastValue.YValue))
	} else {
		lastValue.XValue, lastValue.YValue = innerSeries.GetValues(innerSeries.Len() - 1)
		lastValue.Label = fmt.Sprintf("%s – %d", label, int(lastValue.YValue))
	}

	var seriesName string
	var seriesStyle chart.Style
	if typed, isTyped := innerSeries.(chart.Series); isTyped {
		seriesName = fmt.Sprintf("%s - Last Value", typed.GetName())
		seriesStyle = typed.GetStyle()
	}

	return chart.AnnotationSeries{
		Name:        seriesName,
		Style:       seriesStyle,
		Annotations: []chart.Value2{lastValue},
	}
}
