package util

import (
	"reflect"
	"testing"
)

func TestMaxMinInt(t *testing.T) {
	arr := []int{7, 12, 1, 5, 8, 12, 9, 6}

	max, min := MaxMinInt(arr)

	if max != 12 || min != 1 {
		t.Fail()
	}
}

func TestMaxMinFloat(t *testing.T) {
	arr := []float64{7, 12, 1, 5, 8, 12, 9, 6}

	max, min := MaxMinFloat(arr)

	if max != 12 || min != 1 {
		t.Fail()
	}
}

func TestStripUsernames(t *testing.T) {
	arr := []string{"<@!99201752280096768>", "99201752280096768>", "99201752280096768"}
	results := StripUsernames(arr)
	expected := []string{"99201752280096768", "99201752280096768", "99201752280096768"}

	for i := range results {
		if expected[i] != results[i] {
			t.Error("Expected", expected[i], "got", results[i])
		}
	}
}

func TestSpaceNormalizer(t *testing.T) {
	t.Run("", func(t *testing.T) {
		args := SpaceNormalizer("")
		expected := ""

		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Got %v,\nexpected %v.", args, expected)
		}
	})
	t.Run("1+2 spaces", func(t *testing.T) {
		args := SpaceNormalizer("1  ")
		expected := "1"

		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Got %v,\nexpected %v.", args, expected)
		}
	})
	t.Run("1+2 spaces before", func(t *testing.T) {
		args := SpaceNormalizer("  1")
		expected := "1"

		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Got %v,\nexpected %v.", args, expected)
		}
	})
	t.Run("1+2 spaces on each side", func(t *testing.T) {
		args := SpaceNormalizer("  1  ")
		expected := "1"

		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Got %v,\nexpected %v.", args, expected)
		}
	})
	t.Run("sub w/ no spaces", func(t *testing.T) {
		args := SpaceNormalizer("sub<@!99201752280096768><@!99201752280096768>")
		expected := "sub <@!99201752280096768> <@!99201752280096768>"

		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Got %v,\nexpected %v.", args, expected)
		}
	})
	t.Run("sub w/ too many spaces", func(t *testing.T) {
		args := SpaceNormalizer("sub  <@!99201752280096768>  <@!99201752280096768>  ")
		expected := "sub <@!99201752280096768> <@!99201752280096768>"

		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Got %v,\nexpected %v.", args, expected)
		}
	})
	t.Run("sub w/ no spaces between", func(t *testing.T) {
		args := SpaceNormalizer("sub <@!99201752280096768><@!99201752280096768>")
		expected := "sub <@!99201752280096768> <@!99201752280096768>"

		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Got %v,\nexpected %v.", args, expected)
		}
	})
	t.Run("sub correctly", func(t *testing.T) {
		args := SpaceNormalizer("sub <@!99201752280096768> <@!99201752280096768>")
		expected := "sub <@!99201752280096768> <@!99201752280096768>"

		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Got %v,\nexpected %v.", args, expected)
		}
	})
}
