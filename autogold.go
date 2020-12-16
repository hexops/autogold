package autogold

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/gotextdiff"
)

var (
	update       = flag.Bool("update", false, "update .golden files, removing unused if running all tests")
	updateOnly   = flag.Bool("update-only", false, "update .golden files, leaving unused")
	noUpdateFail = flag.Bool("no-update-fail", false, "do not fail tests if .golden file was updated")
	cleaned      = map[string]struct{}{}
	cleanDir     string
)

func shouldUpdateOnly() bool {
	if *updateOnly {
		return true
	}
	if *update {
		for _, arg := range os.Args {
			if strings.HasPrefix(arg, "-test.run") {
				// Running a subset of the tests, so don't remove unused files.
				return true
			}
		}
	}
	return false
}

func mkTempDir() error {
	if cleanDir != "" {
		return nil
	}

	// Try to remove past go-golden temp dirs.
	matches, err := filepath.Glob(filepath.Join(os.TempDir(), "go-golden-*"))
	if err != nil {
		return err
	}
	for _, match := range matches {
		if err := os.RemoveAll(match); err != nil {
			return err
		}
	}

	// Create a temp dir for this run.
	cleanDir, err = ioutil.TempDir("", "go-golden-*")
	if err != nil {
		return err
	}
	return nil
}

// Equal checks if got is equal to the saved testdata/.golden test file. If it is not, t.Fatal
// is called with a multi-line diff comparison.
//
// If the go test -update flag is specified, the .golden files will be updated or created automatically.
//
// Custom equality operators can be used if needed by passing options. See https://pkg.go.dev/github.com/google/go-cmp/cmp
func Equal(t *testing.T, got interface{}, opts ...Option) {
	var (
		cmpOpts  []cmp.Option
		dir      = "testdata"
		fileName = strings.Replace(t.Name(), "/", "__", -1)
	)
	for _, opt := range opts {
		opt := opt.(*option)
		if opt.name != "" {
			fileName = opt.name
		}
		if opt.opts != nil {
			cmpOpts = opt.opts
		}
		if opt.dir != "" {
			dir = opt.dir
		}
	}

	outFile := filepath.Join(dir, fileName+".golden")

	if !shouldUpdateOnly() {
		if err := mkTempDir(); err != nil {
			t.Fatal(err)
		}
		_, ok := cleaned[dir]
		if !ok {
			// Move all .golden files in the directory into the temp dir.
			cleaned[dir] = struct{}{}
			matches, err := filepath.Glob(filepath.Join(dir, "*.golden"))
			if err != nil {
				t.Fatal(err)
			}
			for _, match := range matches {
				err := os.Rename(match, filepath.Join(cleanDir, filepath.Base(match)))
				if err != nil {
					t.Fatal(err)
				}
			}
		}

		// Move the golden file for this test back into the testdata dir, if it exists.
		tmpFile := filepath.Join(cleanDir, filepath.Base(fileName+".golden"))
		err := os.Rename(tmpFile, outFile)
		if err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}

	want, err := ioutil.ReadFile(outFile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}

	gotString := stringify(got, cmpOpts...)
	edits := myers.ComputeEdits(span.URIFromPath("out"), string(want), gotString)
	diff := fmt.Sprint(gotextdiff.ToUnified("want", "got", string(want), edits))
	if diff != "" {
		if *update || shouldUpdateOnly() {
			outDir := filepath.Dir(outFile)
			if _, err := os.Stat(outDir); os.IsNotExist(err) {
				if err := os.MkdirAll(outDir, 0700); err != nil {
					t.Fatal(err)
				}
			}
			if err := ioutil.WriteFile(outFile, []byte(gotString), 0666); err != nil {
				t.Fatal(err)
			}
		}
		if !*noUpdateFail {
			t.Fatal(fmt.Errorf("mismatch (-want +got):\n%s", diff))
		}
	}
}

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

func stringify(v interface{}, opts ...cmp.Option) string {
	if v, ok := v.(string); ok {
		return v
	}
	if v, ok := v.(fmt.GoStringer); ok {
		return v.GoString()
	}
	reporter := &stringReporter{}
	cmp.Equal(v, v, append(opts, cmp.Reporter(reporter))...)
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
