package autogold

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

func diff(got, want string, opts []Option) string {
	edits := myers.ComputeEdits(span.URIFromPath("out"), string(want), got)
	return fmt.Sprint(gotextdiff.ToUnified("want", "got", string(want), edits))
}

func stringify(v interface{}, opts []Option) string {
	var cmpOpts []cmp.Option
	for _, opt := range opts {
		opt := opt.(*option)
		if opt.opts != nil {
			cmpOpts = opt.opts
		}
	}
	if v, ok := v.(string); ok {
		return v
	}
	if v, ok := v.(fmt.GoStringer); ok {
		return v.GoString()
	}
	reporter := &stringReporter{}
	cmp.Equal(v, v, append(cmpOpts, cmp.Reporter(reporter))...)
	return reporter.String() + "\n"
}

// stringReporter implements the cmp.Reporter interface by writing out the entire object as a
// string.
type stringReporter struct {
	path   cmp.Path
	fields []string
}

func (s *stringReporter) PushStep(ps cmp.PathStep) {
	s.path = append(s.path, ps)
}

func (s *stringReporter) Report(rs cmp.Result) {
	vx, _ := s.path.Last().Values()
	var v string
	switch {
	case vx.Kind() == reflect.String:
		v = fmt.Sprintf("%q", vx.String())
	default:
		v = fmt.Sprintf("%+v", vx)
	}
	s.fields = append(s.fields, fmt.Sprintf("%#v:%s", s.path, v))
}

func (s *stringReporter) PopStep() {
	s.path = s.path[:len(s.path)-1]
}

func (s *stringReporter) String() string {
	return strings.Join(s.fields, "\n")
}
