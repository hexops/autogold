package example

type Baz struct {
	Name string
	Age  int
}

func Bar() *Baz {
	return &Baz{
		Name: "Jane",
		Age:  31,
	}
}
