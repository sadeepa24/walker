# Walker

Walker is a runtime structure modifier that allows dynamic changes to any struct in your application using reflection. It provides a powerful and flexible way to manipulate data structures at runtime, enabling advanced use cases such as dynamic configuration, testing, and more.

## Features

- Modify struct fields dynamically at runtime.
- Leverage Go's reflection capabilities for advanced struct manipulation.
- Lightweight and easy to integrate into existing projects.

## Installation

To install Walker, use the following command:

```bash
go get github.com/sadeepa24/walker
```

## Usage

Here's a quick example of how to use Walker:

```go
package main

import (
    "fmt"
    "log"
    "github.com/sadeepa24/walker"
)

type ExampleStruct struct {
    Name  string
    Age   int
    Email string
}

func main() {
	obj := ExampleStruct{Name: "John", Age: 30, Email: "john@example.com"}

	fmt.Println("Before:", obj)
	wlkr, err := walker.NewWalker(&obj)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("items in object", wlkr.Items())

	if walked := wlkr.WalkInto("Name"); walked {
		wlkr.Change("new name")
		wlkr.WalkBack()
	}
	if walked := wlkr.WalkInto("Age"); walked {
		wlkr.Change(20)
		wlkr.WalkBack()
	}
	fmt.Println("After:", obj)
}
```

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

### Runtime Structure Modifier Using Reflect
