package bot

import (
	"reflect"
	"testing"
)

func TestParseGraphArgs(t *testing.T) {
	t.Run("g!", func(t *testing.T) {
		args, err := ParseGraphArgs("g!")
		expected := Args{Top: 10}

		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Got %v,\nexpected %v.\nError: %s", args, expected, err.Error())
		}
	})

	t.Run("g! top 5", func(t *testing.T) {
		args, err := ParseGraphArgs("g! top 5")
		expected := Args{Top: 5}

		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Got %v,\nexpected %v.\nError: %s", args, expected, err.Error())
		}
	})

	t.Run("g! users Crouton ode", func(t *testing.T) {
		args, err := ParseGraphArgs("g! users Crouton ode")
		expected := Args{Usernames: []string{"Crouton", "ode"}}

		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Got %v,\nexpected %v.\nError: %s", args, expected, err.Error())
		}
	})

	t.Run("g! king", func(t *testing.T) {
		args, err := ParseGraphArgs("g! king")
		expected := Args{Top: 10, King: true}

		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Got %v,\nexpected %v.\nError: %s", args, expected, err.Error())
		}
	})
}
