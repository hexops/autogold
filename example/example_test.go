package example

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestEqual(t *testing.T) {
	got := Bar()
	autogold.Equal(t, got)
}

func TestSubtest(t *testing.T) {
	// Create one of these per sub-test value you want to compare.
	want := autogold.Want("mysubtest", &Baz{Name: "Jane", Age: 31})

	// Invoke test.Equal once you have your result.
	t.Run(want.Name(), func(t *testing.T) {
		got := Bar()
		want.Equal(t, got)
	})
}
