# autogold - automated Go golden file testing <a href="https://hexops.com"><img align="right" alt="Hexops logo" src="https://raw.githubusercontent.com/hexops/media/master/readme.svg"></img></a>

<a href="https://pkg.go.dev/badge/github.com/hexops/autogold"><img src="https://pkg.go.dev/badge/badge/github.com/hexops/autogold.svg" alt="Go Reference" align="right"></a>

Instead of writing your desired test output as a large Go structure / string in your code, simply write:

```Go
autogold.Equal(t, got)
```

The test output (nested Go struct, string, etc.) will be formatted using [valast](https://github.com/hexops/valast). If the `testdata/<test name>.golden` snapshot file is different, the test will fail with a [nice multi-line diff](https://github.com/hexops/gotextdiff) and `go test -update` will update the file if you like the changes.

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
        &example.Baz{Name: "Jane", Age: 31}
```

We see a diff showing what our test produced and what we expected (nothing, because `testdata/TestFoobar.golden` does not exist.)

Rerun the test with `go test -update` and `testdata/TestFoobar.golden` will be created/updated with the output we got.

## When should golden files be used?

Golden files are used by the Go authors for testing [the standard library](https://golang.org/src/go/doc/doc_test.go), the [`gofmt` tool](https://github.com/golang/go/blob/master/src/cmd/gofmt/gofmt_test.go#L124-L130), etc. and are a common pattern in the Go community for snapshot testing. See also ["Testing with golden files in Go" - Chris Reeves](https://medium.com/soon-london/testing-with-golden-files-in-go-7fccc71c43d3)

_Golden files make the most sense when you'd otherwise have to write a complex multi-line string or large Go structure inline in your test._

## Custom formatting

[valast](https://github.com/hexops/valast) is used to produce a text description of the Go value you provide to `autogold.Equal`. If the default output is not suitable for your test, you have options:

### Changing formatting for a specific sub-value

If the value you provide implements the [`fmt.GoStringer`](https://golang.org/pkg/fmt/#GoStringer) interface, it will be used to convert your type to a string.

### Exclude exported fields

```Go
autogold.Equal(t, got, autogold.ExportedOnly())
```

### Use your own custom format (JSON, etc.)

Simply pass a `string` value to `autogold.Equal`, doing the formatting yourself. It'll be written to the golden file as-is.

## Alternatives

The following are alternatives to autogold:

- [github.com/xorcare/golden](https://pkg.go.dev/github.com/xorcare/golden): works on `[]byte` inputs only.
- [github.com/sebdah/goldie](https://pkg.go.dev/github.com/sebdah/goldie/v2): works on `[]byte` inputs only (but provides helpers for JSON, XML, etc.)
- [github.com/bradleyjkemp/cupaloy](https://pkg.go.dev/github.com/bradleyjkemp/cupaloy/v2) works on `interface{}` inputs, but [uses inactive go-spew project](https://github.com/davecgh/go-spew/issues/128) to format Go structs.
