package util

import "testing"

func TestMaxMin(t *testing.T) {
	arr := []int{7, 12, 1, 5, 8, 12, 9, 6}

	max, min := MaxMin(arr)

	if max != 12 || min != 1 {
		t.Fail()
	}
}

func TestStripUsernames(t *testing.T) {
	arr := []string{"@ode#0000", "@Crouton#1234", "@Pupper"}
	results := StripUsernames(arr)
	expected := []string{"ode", "Crouton", "Pupper"}

	for i := range results {
		if expected[i] != results[i] {
			t.Error("Expected", expected[i], "got", results[i])
		}
	}
}
