package autogold

import (
	"github.com/google/go-cmp/cmp"
)

// Option configures specific behavior for Equal.
type Option interface {
	// isValidOption is an unexported field to ensure only valid options from this package can be
	// used.
	isValidOption()
}

type option struct {
	name string
	opts []cmp.Option
	dir  string
}

func (o *option) isValidOption() {}

// Unexported is an option that includes unexported fields in the output.
func Unexported() Option {
	return CmpOptions(cmp.AllowUnexported())
}

// CmpOptions specifies go-cmp options that should be used.
func CmpOptions(opts ...cmp.Option) Option {
	return &option{opts: opts}
}

// Name specifies a name to use for the testdata/<name>.golden file instead of the default test name.
func Name(name string) Option {
	return &option{name: name}
}

// Dir specifies a customer directory to use for writing the golden files, instead of the default "testdata/".
func Dir(dir string) Option {
	return &option{dir: dir}
}
