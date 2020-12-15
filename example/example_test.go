package example

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestFoo(t *testing.T) {
	got := Bar()
	autogold.Equal(t, got)
}
