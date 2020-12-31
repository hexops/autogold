package autogold

import (
	"fmt"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/hexops/valast"
)

func diff(got, want string, opts []Option) string {
	edits := myers.ComputeEdits(span.URIFromPath("out"), string(want), got)
	return fmt.Sprint(gotextdiff.ToUnified("want", "got", string(want), edits))
}

func stringify(v interface{}, opts []Option) string {
	if v, ok := v.(string); ok {
		return v
	}
	if v, ok := v.(fmt.GoStringer); ok {
		return v.GoString()
	}
	var valastOpt *valast.Options
	for _, opt := range opts {
		opt := opt.(*option)
		if opt.exportedOnly {
			valastOpt.ExportedOnly = true
		}
	}
	return valast.StringWithOptions(v, valastOpt) + "\n"
}
