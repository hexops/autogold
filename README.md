# autogold - automatically update your Go tests <a href="https://hexops.com"><img align="right" alt="Hexops logo" src="https://raw.githubusercontent.com/hexops/media/master/readme.svg"></img></a>

<a href="https://pkg.go.dev/badge/github.com/hexops/autogold"><img src="https://pkg.go.dev/badge/badge/github.com/hexops/autogold.svg" alt="Go Reference" align="right"></a>

autogold makes `go test -update` automatically update your Go tests (golden files and Go values in e.g. `foo_test.go`.)

## Automatic golden files

Write in a Go test:

```Go
import "github.com/hexops/autogold"
...
autogold.Equal(t, got)
```

`go test -update` will now create/update a `testdata/<test name>.golden` file for you automatically.

## Automatic inline test updating

Write in a Go test:

```Go
autogold.Inline(t, got, nil)
```

`go test -update` will automatically update `nil` with the Go syntax for whatever value your test `got` (complex Go struct, slices, strings, etc.)

## Diffs

Anytime your test produces a result that is unexpected, you'll get very nice diffs showing exactly what changed. It does this by [converting values at runtime directly to a formatted Go AST](https://github.com/hexops/valast), and using the same [diffing library the Go language server uses](https://github.com/hexops/gotextdiff):

```

```

## Subtesting

Use [table-driven Go subtests](https://blog.golang.org/subtests)? `autogold.Want` and `go test -update` will automatically find and replace the `nil` values for you:

```Go
func TestTime(t *testing.T) {
	testCases := []struct {
		gmt  string
		loc  string
		want autogold.Value
	}{
		{"12:31", "Europe/Zuri", autogold.Want("Europe", nil)},
		{"12:31", "America/New_York", autogold.Want("America", nil)},
		{"08:08", "Australia/Sydney", autogold.Want("Australia", nil)},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			loc, err := time.LoadLocation(tc.loc)
			if err != nil {
				t.Fatal("could not load location")
			}
			gmt, _ := time.Parse("15:04", tc.gmt)
			got := gmt.In(loc).Format("15:04")
			tc.want.Equal(t, got)
		})
	}
}
```

It works by finding the relevant `autogold.Want("<unique name>", ...)` call below the named `TestTime` function, and then replacing the `nil` parameter (or anything that was there.)

## What are golden files, when should they be used?

Golden files are used by the Go authors for testing [the standard library](https://golang.org/src/go/doc/doc_test.go), the [`gofmt` tool](https://github.com/golang/go/blob/master/src/cmd/gofmt/gofmt_test.go#L124-L130), etc. and are a common pattern in the Go community for snapshot testing. See also ["Testing with golden files in Go" - Chris Reeves](https://medium.com/soon-london/testing-with-golden-files-in-go-7fccc71c43d3)

_Golden files make the most sense when you'd otherwise have to write a complex multi-line string or large Go structure inline in your test, making it hard to read._

In most cases, you should prefer inline snapshots, subtest golden values, or traditional Go tests.

## Custom formatting

[valast](https://github.com/hexops/valast) is used to produce Go syntax at runtime for the Go value you provide. If the default output is not to your liking, you have options:

- **Pass a string to autogold**: It will be formatted as a Go string for you in the resulting `.golden` file / in Go tests.
- **Use your own formatting (JSON, etc.)**: Make your `got` value of type `autogold.Raw("foobar")`, and it will be used as-is for `.golden` files (not allowed with inline tests.)
- **Exclude unexported fields**: `autogold.Equal(t, got, autogold.ExportedOnly())`

## Alternatives comparison

The following are alternatives to autogold, making note of the differences we found that let us to create autogold:

- [github.com/xorcare/golden](https://pkg.go.dev/github.com/xorcare/golden)
    - Supports `[]byte` inputs only, defers formatting to users.
    - Does not support inline snapshots / code updating.
- [github.com/sebdah/goldie](https://pkg.go.dev/github.com/sebdah/goldie/v2)
    - Supports `[]byte` inputs only, provides helpers for JSON, XML, etc.
    - Does not support inline snapshots / code updating.
- [github.com/bradleyjkemp/cupaloy](https://pkg.go.dev/github.com/bradleyjkemp/cupaloy/v2)
    - Works on `interface{}` inputs.
    - [Uses inactive go-spew project](https://github.com/davecgh/go-spew/issues/128) to format Go structs.
    - Does not support inline snapshots / code updating.
