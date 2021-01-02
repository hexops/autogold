# autogold - automatically update your Go tests <a href="https://hexops.com"><img align="right" alt="Hexops logo" src="https://raw.githubusercontent.com/hexops/media/master/readme.svg"></img></a>

<a href="https://pkg.go.dev/badge/github.com/hexops/autogold"><img src="https://pkg.go.dev/badge/badge/github.com/hexops/autogold.svg" alt="Go Reference" align="right"></a>

autogold makes `go test -update` automatically update your Go tests. It can automatically create/update:

- [`testdata/.golden` file tests](#golden-file-testing)
- [Inline snapshots / golden values](#inline-snapshots-golden-values) (including complex Go structs, etc.)
- [Golden subtest values](#golden-subtest-values)

## Golden file testing

Write in a Go test:

```Go
import "github.com/hexops/autogold"
...
autogold.File(t, got)
```

`go test -update` will automatically create/update a `testdata/<test name>.golden` file for you.

## Inline snapshots / golden values

Write in a Go test:

```Go
autogold.Inline(t, got, nil)
```

`go test -update` will automatically update `nil` with the Go syntax for whatever value your test `got` (complex Go struct, slices, strings, etc.)

## Golden subtest values

Use [table-driven Go subtests](https://blog.golang.org/subtests)? `autogold.Value` and `go test -update` will automatically find and replace the `nil` values here for you:

```Go
func TestTime(t *testing.T) {
    testCases := []struct {
        gmt  string
        loc  string
        test autogold.Test
    }{
        {"12:31", "Europe/Zuri", autogold.Value("Europe", nil)},
        {"12:31", "America/New_York", autogold.Value("America", nil)},
        {"08:08", "Australia/Sydney", autogold.Value("Australia", nil)},
    }
    for _, tc := range testCases {
        t.Run(tc.test.Name(), func(t *testing.T) {
            loc, err := time.LoadLocation(tc.loc)
            if err != nil {
                t.Fatal("could not load location")
            }
            gmt, _ := time.Parse("15:04", tc.gmt)
            got := gmt.In(loc).Format("15:04")
            tc.test.Check(t, got)
        })
    }
}
```

It does this by looking for the relevant`autogold.Value("<subtest name>", ...)` call below the named test, and replacing the `nil` parameter.

## What are golden files, when should they be used?

Golden files are used by the Go authors for testing [the standard library](https://golang.org/src/go/doc/doc_test.go), the [`gofmt` tool](https://github.com/golang/go/blob/master/src/cmd/gofmt/gofmt_test.go#L124-L130), etc. and are a common pattern in the Go community for snapshot testing. See also ["Testing with golden files in Go" - Chris Reeves](https://medium.com/soon-london/testing-with-golden-files-in-go-7fccc71c43d3)

_Golden files make the most sense when you'd otherwise have to write a complex multi-line string or large Go structure inline in your test, making it hard to read._

In most cases, you should prefer inline snapshots, subtest golden values, or traditional Go tests.

## Custom formatting

[valast](https://github.com/hexops/valast) is used to produce Go syntax at runtime for the Go value you provide to `autogold`. If the default output is not suitable for your test, you have options:

- **Use your own formatting (JSON, etc.)**: Make your `got` value of type `autogold.Raw("foobar")`, and it will be used as-is.
- **Exclude unexported fields**: `autogold.Foo(t, got, autogold.ExportedOnly())`

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
