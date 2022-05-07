package example

type ExampleFull struct {
	Field string
}

func (ExampleFull) Foo() string {
	return "42"
}
