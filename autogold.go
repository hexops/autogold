package autogold

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/fatih/color"
	"github.com/nightlyone/lockfile"
)

var (
	update       = flag.Bool("update", false, "update .golden files, removing unused if running all tests")
	updateOnly   = flag.Bool("update-only", false, "update .golden files, leaving unused")
	noUpdateFail = flag.Bool("no-update-fail", false, "do not fail tests if .golden file was updated")
	cleaned      = map[string]struct{}{}
	cleanDir     string
)

func init() {
	color.NoColor = false
}

// Equal checks if got is equal to the saved `testdata/<test name>.golden` test file. If it is not,
// t.Fatal is called with a multi-line diff comparison.
//
// If the `go test -update` flag is specified, the .golden files will be updated/created
// automatically.
//
// If the input value is of type Raw, its contents will be directly used instead of the value being
// formatted as a Go literal.
func Equal(t *testing.T, got interface{}, opts ...Option) {
	dir := testdataDir(opts)
	fileName := testName(t, opts)
	outFile := filepath.Join(dir, fileName+".golden")

	// At this point dir may be "testdata/" while outFile may be "testdata/TestFoo/subTest.golden".
	// Reconcile this situation so we can rely on dir for e.g. removing unused .golden files in it,
	// locking it (instead of the entire "testdata/" directory), etc.
	dir = filepath.Dir(outFile)

	// grabLock will acquire a directory-level lock to prevent concurrent mutations to the .golden
	// files by parallel tests (whether in-process, or not.)
	var goldenFilesUnlock func() error
	grabLock := func() {
		if goldenFilesUnlock != nil {
			return
		}
		var err error
		goldenFilesUnlock, err = acquirePathLock(dir)
		if err != nil {
			t.Fatal(err)
		}
	}
	unlock := func() {
		if goldenFilesUnlock != nil {
			if err := goldenFilesUnlock(); err != nil {
				t.Fatal(err)
			}
			goldenFilesUnlock = nil
		}
	}
	defer unlock()

	if !shouldUpdateOnly() && *update {
		if err := mkTempDir(); err != nil {
			t.Fatal(err)
		}
		grabLock()
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
		unlock() // don't hold the lock while we perform IO, diffing, etc. below.
	}

	want, err := ioutil.ReadFile(outFile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}

	opts = append(opts, &option{allowRaw: true, trailingNewline: true})
	gotString := stringify(got, opts)
	diff := diff(gotString, string(want), opts)

	_, isRaw := got.(Raw)
	isEmptyFile := isRaw && gotString == ""
	if isEmptyFile && (*update || shouldUpdateOnly()) {
		grabLock()
		os.Remove(outFile)
	}
	if diff != "" {
		if *update || shouldUpdateOnly() {
			grabLock()
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			if err := ioutil.WriteFile(outFile, []byte(gotString), 0o666); err != nil {
				t.Fatal(err)
			}
		}
		if !*noUpdateFail {
			t.Fatal(fmt.Errorf("mismatch (-want +got):\n%s", colorDiff(diff)))
		}
	}
}

func colorDiff(diff string) string {
	s := []string{}

	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "-") {
			s = append(s, color.RedString(line[1:]))
		} else if strings.HasPrefix(line, "+") {
			s = append(s, color.GreenString(line[1:]))
		} else if strings.HasPrefix(line, " ") {
			s = append(s, line[1:])
		} else {
			s = append(s, line)
		}
	}

	return strings.Join(s, "\n")
}

var (
	pathLocksMu sync.Mutex
	pathLocks   = map[string]*pathLock{}
)

type pathLock struct {
	ownership sync.Mutex
	lockfile  lockfile.Lockfile
}

// acquirePathLock acquires a PID-based lockfile for the given path, which will be made into an
// absolute path.
//
// The returned function unlocks the lock.
func acquirePathLock(path string) (func() error, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	sha := fmt.Sprintf("%x", sha256.Sum256([]byte(path)))
	pathHash := string(sha[:7])
	lockPath := filepath.Join(os.TempDir(), "autogold."+pathHash)

	pathLocksMu.Lock()
	lock, inProcessAlready := pathLocks[lockPath]
	if !inProcessAlready {
		lockfile, err := lockfile.New(lockPath)
		if err != nil {
			pathLocksMu.Unlock()
			return nil, err
		}
		lock = &pathLock{lockfile: lockfile}
		pathLocks[lockPath] = lock
	}
	pathLocksMu.Unlock()

	// Must not have multiple goroutines own the lockfile.
	lock.ownership.Lock()
	if err := lock.lockfile.TryLock(); err != nil {
		lock.ownership.Unlock()
		return nil, err
	}
	return func() error {
		defer lock.ownership.Unlock()
		if err := lock.lockfile.Unlock(); err != nil {
			return fmt.Errorf("failed to unlock %q, reason: %v (you may need to delete the file)", lock.lockfile, err)
		}
		return nil
	}, nil
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
