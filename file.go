package autogold

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	update       = flag.Bool("update", false, "update .golden files, removing unused if running all tests")
	updateOnly   = flag.Bool("update-only", false, "update .golden files, leaving unused")
	noUpdateFail = flag.Bool("no-update-fail", false, "do not fail tests if .golden file was updated")
	cleaned      = map[string]struct{}{}
	cleanDir     string
)

// Equal checks if got is equal to the saved `testdata/<test name>.golden` test file. If it is not,
// t.Fatal is called with a multi-line diff comparison.
//
// If the `go test -update` flag is specified, the .golden files will be updated/created
// automatically.
func Equal(t *testing.T, got interface{}, opts ...Option) {
	dir := testdataDir(opts)
	fileName := strings.Replace(testName(t, opts), "/", "__", -1)
	outFile := filepath.Join(dir, fileName+".golden")

	if !shouldUpdateOnly() && *update {
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

	gotString := stringify(got, opts) + "\n"
	diff := diff(gotString, string(want), opts)
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

func testName(t *testing.T, opts []Option) string {
	for _, opt := range opts {
		opt := opt.(*option)
		if opt.name != "" {
			return opt.name
		}
	}
	return t.Name()
}

func testdataDir(opts []Option) string {
	for _, opt := range opts {
		opt := opt.(*option)
		if opt.dir != "" {
			return opt.dir
		}
	}
	return "testdata"
}
