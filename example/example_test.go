package example

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestExpectFile(t *testing.T) {
	got := Bar()
	autogold.ExpectFile(t, got)
}

func TestInline(t *testing.T) {
	got := Bar()
	// First parameter to Expect can be omitted since we're only using a single Expect in this test.
	//
	// 2nd parameter to Expect is the wanted value - autogold will update this for us.
	autogold.Expect(&Baz{Name: "Jane", Age: 31}).Equal(t, got)
}

func TestSubtest(t *testing.T) {
	// Create one of these per sub-test value you want to compare.
	expect := autogold.Expect(&Baz{Name: "Jane", Age: 31})

	// Invoke test.Equal once you have your result.
	t.Run("my test", func(t *testing.T) {
		got := Bar()
		expect.Equal(t, got)
	})
}
