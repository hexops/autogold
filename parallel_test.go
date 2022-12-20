package autogold_test

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestWant_parallel1(t *testing.T) {
	testParallel(t, "1")
}

func TestWant_parallel2(t *testing.T) {
	testParallel(t, "2")
}

func testParallel(t *testing.T, prefix string) {
	t.Parallel()

	testTable := []string{
		prefix + "-first",
		prefix + "-second",
	}

	for _, name := range testTable {
		name := name

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			autogold.ExpectFile(t, name)
		})
	}
}
