package autogold

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

// Value describes a desired value for a Go test, see Want for more information.
type Value interface {
	// Name returns the value name.
	Name() string

	// Equal checks if `got` matches the desired test value, invoking t.Fatal otherwise.
	Equal(t *testing.T, got interface{}, opts ...Option)
}

type value struct {
	name  string
	equal func(t *testing.T, got interface{}, opts ...Option)
}

func (v value) Name() string { return v.name }
func (v value) Equal(t *testing.T, got interface{}, opts ...Option) {
	v.equal(t, got, opts...)
}

var (
	getPackageNameAndPathCacheMu sync.RWMutex
	getPackageNameAndPathCache   = map[string][2]string{}
)

func getPackageNameAndPath(dir string) (name, path string, err error) {
	// If it is cached, fetch it from the cache. This prevents us from doing a semi-costly package
	// load for every test that runs, instead requiring we only do it once per _test.go directory.
	getPackageNameAndPathCacheMu.RLock()
	if v, cached := getPackageNameAndPathCache[dir]; cached {
		getPackageNameAndPathCacheMu.RUnlock()
		return v[0], v[1], nil
	}
	getPackageNameAndPathCacheMu.RUnlock()

	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName}, dir)
	if err != nil {
		return "", "", err
	}
	getPackageNameAndPathCacheMu.Lock()
	getPackageNameAndPathCache[dir] = [2]string{pkgs[0].Name, pkgs[0].PkgPath}
	getPackageNameAndPathCacheMu.Unlock()
	return pkgs[0].Name, pkgs[0].PkgPath, nil
}

// Want returns a desired Value which can later be checked for equality against a gotten value.
//
// The name parameter must be a Go string literal (NOT a variable or expression), and must be unique
// within the Go test function.
//
// When `-update` is specified, autogold will find and replace in the test file by looking for an
// instance of e.g. `autogold.Want("bar", ...)` beneath the calling `TestFoo` function and replacing
// the `want` value parameter.
func Want(name string, want interface{}) Value {
	return value{
		name: name,
		equal: func(t *testing.T, got interface{}, opts ...Option) {
			// Identify the root test name ("TestFoo" in "TestFoo/bar")
			testName := t.Name()
			if strings.Contains(testName, "/") {
				split := strings.Split(testName, "/")
				testName = split[0]
			}

			// Find the path to the calling _test.go, relative to where the test is being run.
			var (
				file string
				ok   bool
			)
			for caller := 1; ; caller++ {
				_, file, _, ok = runtime.Caller(caller)
				if !ok || strings.Contains(file, "_test.go") {
					break
				}
			}
			if !ok {
				t.Fatal("runtime.Caller: returned ok=false")
			}
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			testPath, err := filepath.Rel(pwd, file)
			if err != nil {
				t.Fatal(err)
			}

			// Determine the package name and path of the test file, so we can unqualify types in
			// that package.
			pkgName, pkgPath, err := getPackageNameAndPath(filepath.Dir(testPath))
			if err != nil {
				t.Fatalf("loading package: %v", err)
			}
			opts = append(opts, &option{
				forPackagePath: pkgPath,
				forPackageName: pkgName,
			})

			// Check if the test failed or not by diffing the results.
			wantString := stringify(want, opts)
			gotString := stringify(got, opts)
			diff := diff(gotString, wantString, opts)
			if diff == "" {
				return // test passed
			}

			// Update the test file if so desired.
			if *update || shouldUpdateOnly() {
				// Acquire a file-level lock to prevent concurrent mutations to the _test.go file
				// by parallel tests (whether in-process, or not.)
				unlock, err := acquirePathLock(testPath)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := unlock(); err != nil {
						t.Fatal(err)
					}
				}()

				// Replace the autogold.Want(...) call's `want` parameter with the expression for the
				// value we got.
				newTestFile, err := replaceWant(testPath, testName, name, gotString)
				if err != nil {
					t.Fatal(fmt.Errorf("autogold: %v", err))
				}
				info, err := os.Stat(testPath)
				if err != nil {
					t.Fatal(err)
				}
				if err := ioutil.WriteFile(testPath, []byte(newTestFile), info.Mode()); err != nil {
					t.Fatal(err)
				}
			}
			if !*noUpdateFail {
				t.Fatal(fmt.Errorf("mismatch (-want +got):\n%s", diff))
			}
		},
	}
}

// replaceWant replaces the invocation of:
//
// 	autogold.Want("value_name", ...)
//
// With:
//
// 	autogold.Want("value_name", <replacement>)
//
// Underneath a Go testing function named testName, returning an error if it cannot be found.
//
// The returned updated file contents have the specified replacement, with goimports ran over the
// result.
func replaceWant(testFilePath, testName, valueName, replacement string) ([]byte, error) {
	testFileSrc, err := ioutil.ReadFile(testFilePath)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, testFilePath, testFileSrc, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing file: %v", err)
	}

	// Locate the autogold.Want() call expression and perform string replacement on its 2nd
	// argument.
	//
	// We use string replacement instead of direct ast.Expr swapping so as to ensure that we
	// can use gofumpt to format just our generated ast.Expr, and just gofmt for the remainder
	// of the file (i.e. leaving final formatting of the file up to the user without us having to
	// provide an option.) For why it is important that we use gofumpt on our generated ast.Expr,
	// see https://github.com/hexops/valast/pull/4. As for "why gofmt(goimports) and not gofumpt
	// on the final file?", simply because gofmt is a superset of gofumpt and we don't want to make
	// the call of using gofumpt on behalf of the user.
	callExpr, err := findWantCallExpr(fset, f, testName, valueName)
	if err != nil {
		return nil, err
	}
	arg := callExpr.Args[1]
	start := testFileSrc[:fset.Position(arg.Pos()).Offset]
	end := testFileSrc[fset.Position(arg.End()).Offset:]

	newFile := make([]byte, 0, len(testFileSrc))
	newFile = append(newFile, start...)
	newFile = append(newFile, []byte(replacement)...)
	newFile = append(newFile, end...)
	preFormattingFile := newFile
	newFile, err = imports.Process(testFilePath, newFile, nil)
	if err != nil {
		debug, _ := strconv.ParseBool(os.Getenv("AUTOGOLD_DEBUG"))
		if debug {
			fmt.Println("-------------")
			fmt.Println("ERROR FORMATTING FILE:", err)
			fmt.Println("TEST FILE PATH:", testFilePath)
			fmt.Println("CONTENTS:")
			fmt.Println("-------------")
			fmt.Println(string(preFormattingFile))
			fmt.Println("-------------")
		}
		return nil, fmt.Errorf("formatting file: %v", err)
	}
	return newFile, nil
}

func findWantCallExpr(fset *token.FileSet, f *ast.File, testName, valueName string) (*ast.CallExpr, error) {
	var (
		err             error
		foundTestFunc   bool
		foundCallExpr   *ast.CallExpr
		foundValueNames []string
	)
	pre := func(cursor *astutil.Cursor) bool {
		if err != nil {
			return false
		}
		node := cursor.Node()
		if !foundTestFunc {
			if _, ok := node.(*ast.File); ok {
				return true
			}
			if f, ok := node.(*ast.FuncDecl); ok {
				if f.Name.Name == testName {
					foundTestFunc = true
				}
				return true
			}
			return false
		}
		if foundCallExpr != nil {
			return false
		}
		ce, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		se, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if !isWantSelectorExpr(se) {
			return true
		}
		if len(ce.Args) != 2 {
			return true
		}
		valueNameLit, ok := ce.Args[0].(*ast.BasicLit)
		if !ok || valueNameLit.Kind != token.STRING {
			position := fset.Position(ce.Args[0].Pos())
			err = fmt.Errorf("%s: autogold.Want(...) call must start with a Go string literal", position)
			return false
		}
		var val string
		val, err = strconv.Unquote(valueNameLit.Value)
		if err != nil {
			return false
		}
		if val != valueName {
			foundValueNames = append(foundValueNames, valueNameLit.Value)
			return true
		}
		foundCallExpr = ce
		return true
	}
	f = astutil.Apply(f, pre, nil).(*ast.File)
	if err != nil {
		return nil, err
	}
	if !foundTestFunc {
		return nil, fmt.Errorf("%s: could not find test function: %s", fset.File(f.Pos()).Name(), testName)
	}
	if foundCallExpr == nil {
		if len(foundValueNames) > 0 {
			var didFind string
			if len(foundValueNames) > 2 {
				foundValueNames = foundValueNames[:2]
				didFind = strings.Join(foundValueNames, ", ")
				didFind += ", …"
			} else {
				didFind = strings.Join(foundValueNames, ", ")
			}
			return nil, fmt.Errorf("%s: could not find autogold.Want(%q, ...) function call (did find %s)", fset.File(f.Pos()).Name(), valueName, didFind)
		}
		return nil, fmt.Errorf("%s: could not find autogold.Want(%q, ...) function call", fset.File(f.Pos()).Name(), valueName)
	}
	return foundCallExpr, nil
}

func isWantSelectorExpr(v *ast.SelectorExpr) bool {
	if v.Sel.Name != "Want" {
		return false
	}
	ident, ok := v.X.(*ast.Ident)
	if !ok {
		return false
	}
	// TODO: handle renamed import
	return ident.Name == "autogold"
}
