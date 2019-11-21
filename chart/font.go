package chart

import (
	"io/ioutil"

	"github.com/golang/freetype/truetype"
)

// GetFont returns a font at the given path.
func GetFont(path string) (*truetype.Font, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return &truetype.Font{}, err
	}
	f, err := truetype.Parse(b)
	if err != nil {
		return &truetype.Font{}, err
	}
	return f, nil
}
