package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoved(t *testing.T) {
	table := map[string]struct {
		a []string
		b []string
		x []string
	}{
		"Removed All": {
			a: []string{"a", "b"},
			b: []string{},
			x: []string{"a", "b"},
		},
		"No removes (ensure sort)": {
			a: []string{"b", "c", "a"},
			b: []string{"c", "a", "b"},
			x: []string{},
		},
		"Removed first": {
			a: []string{"a", "b"},
			b: []string{"b"},
			x: []string{"a"},
		},
		"Removed last": {
			a: []string{"a", "b"},
			b: []string{"a"},
			x: []string{"b"},
		},
		"Just Dupes and Adds": {
			a: []string{"a", "d"},
			b: []string{"a", "c", "d", "e"},
			x: []string{},
		},
		"Remove middles": {
			a: []string{"a", "b", "c", "d", "e"},
			b: []string{"a", "c", "e"},
			x: []string{"b", "d"},
		},
		"Remove multiple from first and last": {
			a: []string{"a", "b", "c", "d", "e"},
			b: []string{"c"},
			x: []string{"a", "b", "d", "e"},
		},
		"Remove interspersed": {
			a: []string{"a", "b", "c", "d", "e", "f"},
			b: []string{"d", "f", "b"},
			x: []string{"a", "c", "e"},
		},
	}

	for name, tc := range table {
		t.Run(name, func(t *testing.T) {
			tc := tc // Shadow input for defensive parallelization
			t.Parallel()

			ma := list2map(tc.a)
			mb := list2map(tc.b)

			assert.EqualValues(t, tc.x, removedKeys(ma, mb))
		})
	}
}

func list2map(in []string) map[string]interface{} {
	ret := map[string]interface{}{}
	for _, v := range in {
		ret[v] = true
	}

	return ret
}
