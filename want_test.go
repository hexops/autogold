package autogold

import (
	"fmt"
	"path/filepath"
	"testing"
)

func Test_replaceWant(t *testing.T) {
	tests := []struct {
		file                string
		testName, valueName string
		replacement         string
		err                 string
	}{
		{
			file:        "basic",
			testName:    "TestFoo",
			valueName:   "Bar",
			replacement: `"replacement"`,
		},
		{
			file:        "complex",
			testName:    "TestTime",
			valueName:   "America",
			replacement: `"replacement"`,
		},
		{
			file:      "complex",
			testName:  "TestTime",
			valueName: "Europe",
			replacement: `&struct{
A bool
C error
}{A: true, C: errors.New("abc")}`,
		},
		{
			file:        "complex",
			testName:    "TestTime",
			valueName:   "Australia",
			replacement: `"replacement"`,
		},
		{
			file:        "complex",
			testName:    "TestTime",
			valueName:   "WrongValueName",
			replacement: `"replacement"`,
			err:         `testdata/replace_want/complex: could not find autogold.Want("WrongValueName", ...) function call (did find "Europe", "America", …)`,
		},
		{
			file:        "basic",
			testName:    "TestFoo",
			valueName:   "WrongValueNameWithOthers",
			replacement: `"replacement"`,
			err:         `testdata/replace_want/basic: could not find autogold.Want("WrongValueNameWithOthers", ...) function call (did find "Bar")`,
		},
		{
			file:        "complex",
			testName:    "TestWrongName",
			valueName:   "WrongTestName",
			replacement: `"replacement"`,
			err:         `testdata/replace_want/complex: could not find test function: TestWrongName`,
		},
		{
			file:        "missing",
			testName:    "TestFoo",
			valueName:   "Missing",
			replacement: `"replacement"`,
			err:         `testdata/replace_want/missing: could not find autogold.Want("Missing", ...) function call`,
		},
	}
	for _, tst := range tests {
		t.Run(tst.file+"_"+tst.valueName, func(t *testing.T) {
			testFilePath := filepath.Join("testdata/replace_want", tst.file)
			got, err := replaceWant(testFilePath, tst.testName, tst.valueName, tst.replacement)
			if tst.err != "" && tst.err != fmt.Sprint(err) || tst.err == "" && err != nil {
				t.Fatal("\ngot:\n", err, "\nwant:\n", tst.err)
			}
			Equal(t, string(got))
		})
	}
}
