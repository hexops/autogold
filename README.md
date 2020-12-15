# autogold - automated Go golden file testing [![Go Reference](https://pkg.go.dev/badge/github.com/hexops/autogold.svg)](https://pkg.go.dev/github.com/hexops/autogold) <a href="https://hexops.com"><img align="right" alt="Hexops logo" src="https://raw.githubusercontent.com/hexops/media/master/readme.svg"></img></a>

Instead of writing your desired output as a large structure / string inline in your test, simply write:

```Go
autogold.Equal(t, got)
```

The test output (nested struct, string, etc.) will be formatted using [google/go-cmp](https://github.com/google/go-cmp). If the `testdata/<test name>.golden` snapshot file is different, the test will fail with a [nice multi-line diff](https://github.com/hexops/gotextdiff) and `go test -update` will update the file if you like the changes.

## Example usage

Import the package:

```Go
import "github.com/hexops/autogold"
```

Write a test that produces any Go value (`got`):

```Go
func TestFoo(t *testing.T) {
    got := Bar()
    autogold.Equal(t, got)
}
```

Run `go test` and you'll see the test fails:

```
--- FAIL: TestFoo (0.00s)
    autogold.go:148: mismatch (-want +got):
        --- want
        +++ got
        @@ -1 +1,2 @@
        +{*example.Baz}.Name:"Jane"
        +{*example.Baz}.Age:31
```

We see a diff showing what our test produced and what we expected (nothing, because `testdata/TestFoobar.golden` does not exist.)

Rerun the test with `go test -update` and `testdata/TestFoobar.golden` will be created/updated with the output we got.

## When should golden files be used?

Golden files are used by the Go authors for testing [the standard library](https://golang.org/src/go/doc/doc_test.go), the [`gofmt` tool](https://github.com/golang/go/blob/master/src/cmd/gofmt/gofmt_test.go#L124-L130), etc. and are a common pattern in the Go community for snapshot testing. See also ["Testing with golden files in Go" - Chris Reeves](https://medium.com/soon-london/testing-with-golden-files-in-go-7fccc71c43d3)

_Golden files make the most sense when you'd otherwise have to write a complex multi-line string or large Go structure inline in your test._

## Custom formatting

[google/go-cmp](https://github.com/google/go-cmp) is used to produce a text description of the Go value you provide to `autogold.Equal`. If the default output is not suitable for your test, you have options:

### Changing formatting for a specific sub-value

If your type implements the [`fmt.GoStringer`](https://golang.org/pkg/fmt/#GoStringer) interface, it will be used to convert your type to a string.

### Include unexported fields

```Go
autogold.Equal(t, got, autogold.Unexported())
```

### Use your own custom format (JSON, etc.)

Simply pass a `string` value to `autogold.Equal`, doing the formatting yourself. It'll be written to the golden file as-is.

### Provide custom go-cmp options

You can provide [any go-cmp option](https://pkg.go.dev/github.com/google/go-cmp@v0.5.4/cmp#Option) which will affect formatting by providing the `autogold.CmpOptions(...)` option to `autogold.Equal`.

## Alternatives

Before writing autogold, I considered the following alternatives but was left wanting a better API:

- [github.com/xorcare/golden](https://pkg.go.dev/github.com/xorcare/golden): only works on `[]byte` inputs.
- [github.com/sebdah/goldie](https://pkg.go.dev/github.com/sebdah/goldie/v2): doesn't have a minimal API, only works on `[]byte` inputs (but provides helpers for JSON, XML, etc.)
- [github.com/bradleyjkemp/cupaloy](https://pkg.go.dev/github.com/bradleyjkemp/cupaloy/v2) less minimal API, works on any inputs but [uses inactive go-spew project](https://github.com/davecgh/go-spew/issues/128) to format Go structs.
