package example

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestEqual(t *testing.T) {
	got := Bar()
	autogold.Equal(t, got)
}

func TestInline(t *testing.T) {
	got := Bar()
	// First parameter to Want can be omitted since we're only using a single Want in this test.
	//
	// 2nd parameter to Want is the wanted value - autogold will update this for us.
	autogold.Want("", &Baz{Name: "Jane", Age: 31}).Equal(t, got)
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
