package autogold

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func Test_replaceExpect(t *testing.T) {
	tests := []struct {
		file        string
		testName    string
		line        int
		replacement string
		err         string
	}{
		{
			file:        "basic",
			testName:    "TestFoo",
			line:        10,
			replacement: `"replacement"`,
		},
		{
			file:        "complex",
			testName:    "TestTime",
			line:        16,
			replacement: `"replacement"`,
		},
		{
			file:     "complex",
			testName: "TestTime",
			line:     18,
			replacement: `&struct{
A bool
C error
}{A: true, C: errors.New("abc")}`,
		},
		{
			file:        "complex",
			testName:    "TestTime",
			line:        17,
			replacement: `"replacement"`,
		},
		{
			file:        "complex",
			testName:    "TestTime",
			line:        0,
			replacement: `"replacement"`,
			err:         `testdata/replace_expect/complex: could not find autogold.Expect(…) function call on line 0`,
		},
		{
			file:        "basic",
			testName:    "TestFoo",
			line:        0,
			replacement: `"replacement"`,
			err:         `testdata/replace_expect/basic: could not find autogold.Expect(…) function call on line 0`,
		},
		{
			file:        "complex",
			testName:    "TestWrongName",
			line:        0,
			replacement: `"replacement"`,
			err:         `testdata/replace_expect/complex: could not find autogold.Expect(…) function call on line 0`,
		},
		{
			file:        "missing",
			testName:    "TestFoo",
			line:        0,
			replacement: `"replacement"`,
			err:         `testdata/replace_expect/missing: could not find autogold.Expect(…) function call on line 0`,
		},
		{
			file:        "issue7",
			testName:    "TestNewUserStartTestSuite",
			line:        15,
			replacement: `"replacement"`,
		},
	}
	for _, tst := range tests {
		t.Run(tst.file+"_"+fmt.Sprint(tst.line), func(t *testing.T) {
			testFilePath := filepath.Join("testdata/replace_expect", tst.file)
			got, err := replaceExpect(testFilePath, tst.testName, tst.line, tst.replacement)
			if tst.err != "" && tst.err != fmt.Sprint(err) || tst.err == "" && err != nil {
				t.Fatal("\ngot:\n", err, "\nwant:\n", tst.err)
			}
			Equal(t, Raw(got))
		})
	}
}

func Test_getPackageNameAndPath(t *testing.T) {
	pkgName, pkgPath, err := getPackageNameAndPath(".")
	if err != nil {
		t.Fatal(err)
	}
	if want := "autogold"; pkgName != want {
		t.Fatal("\ngot:\n", pkgName, "\nwant:\n", want)
	}
	if want := "github.com/hexops/autogold"; pkgPath != want {
		t.Fatal("\ngot:\n", pkgPath, "\nwant:\n", want)
	}
}

func Test_getPackageNameAndPath_subdir(t *testing.T) {
	pkgName, pkgPath, err := getPackageNameAndPath("./internal/test")
	if err != nil {
		t.Fatal(err)
	}
	if want := "test"; pkgName != want {
		t.Fatal("\ngot:\n", pkgName, "\nwant:\n", want)
	}
	if want := "github.com/hexops/autogold/internal/test"; pkgPath != want {
		t.Fatal("\ngot:\n", pkgPath, "\nwant:\n", want)
	}
}

func TestEqual_subtestSameNames1(t *testing.T) {
	testEqualSubtestSameNames(t)
}

func TestEqual_subtestSameNames2(t *testing.T) {
	testEqualSubtestSameNames(t)
}

func testEqualSubtestSameNames(t *testing.T) {
	t.Parallel()

	parent := t.Name()

	testTable := []string{
		"first",
		"second",
		"third",
	}

	for _, name := range testTable {
		name := name

		t.Run(name, func(t *testing.T) {
			// Subtests are intentionally not run in parallel, as that makes this issue more easily reproducible

			Equal(t, fmt.Sprintf("%s :: %s", parent, name))
		})
	}
}

func Benchmark_getPackageNameAndPath_cached(b *testing.B) {
	// Wipe the cache, as it was populated by other tests.
	getPackageNameAndPathCacheMu.Lock()
	getPackageNameAndPathCache = map[string][2]string{}
	getPackageNameAndPathCacheMu.Unlock()

	start := time.Now()
	for n := 0; n < b.N; n++ {
		_, _, err := getPackageNameAndPath("./autogold/internal/test")
		if err != nil {
			b.Fatal(err)
		}
		if n == 0 {
			b.Log("first lookup", time.Since(start))
		}
	}
}