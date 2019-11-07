package chart

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestMake(t *testing.T) {
	src := LineChartSource{
		X: []time.Time{
			time.Now().AddDate(0, 0, -1),
			time.Now().AddDate(0, 0, -2),
			time.Now().AddDate(0, 0, -3),
			time.Now().AddDate(0, 0, -4)},
		Series:   [][]float64{[]float64{1, 2, 3, 4}, []float64{1, 2, 8, 9}},
		Labels:   []string{"ode", "Crouton"},
		Title:    "Test",
		LogScale: true,
		Max:      9,
		Min:      1,
	}
	// Create the image
	b, _ := src.Make()
	err := ioutil.WriteFile(os.Getenv("GOPATH")+"/src/github.com/snyderks/xp-bot/lineChart_test.png", b, 0644)
	if err != nil {
		t.Error(err)
	}
}
