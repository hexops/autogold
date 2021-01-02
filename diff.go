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
	// TODO: add Raw type
	if v, ok := v.(string); ok {
		return v
	}
	if v, ok := v.(fmt.GoStringer); ok {
		return v.GoString()
	}
	valastOpt := &valast.Options{}
	for _, opt := range opts {
		opt := opt.(*option)
		if opt.exportedOnly {
			valastOpt.ExportedOnly = true
		}
		if opt.forPackageName != "" {
			valastOpt.PackageName = opt.forPackageName
		}
		if opt.forPackagePath != "" {
			valastOpt.PackagePath = opt.forPackagePath
		}
	}
	return valast.StringWithOptions(v, valastOpt)
}
