Agnostic Raw Data (ARD) for Go
==============================

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Reference](https://pkg.go.dev/badge/github.com/tliron/go-ard.svg)](https://pkg.go.dev/github.com/tliron/go-ard)
[![Go Report Card](https://goreportcard.com/badge/github.com/tliron/go-ard)](https://goreportcard.com/report/github.com/tliron/go-ard)

A library to work with non-schematic data and consume it from various standard formats.

What is ARD? See [here](ARD.md).

This library is [also implemented in Python](https://github.com/tliron/python-ard).

And check out the [ardconv](https://github.com/tliron/ardconv) ARD conversion tool.

Features
--------

Read ARD from any Go `Reader` or decode from strings:

```go
import (
	"fmt"
	"strings"
	"github.com/tliron/go-ard"
)

var yamlRepresentation = `
first:
  property1: Hello
  property2: 1.2
  property3:
  - 1
  - 2
  - second:
      property1: World
`

func main() {
	if value, _, err := ard.Read(strings.NewReader(yamlRepresentation), "yaml", false); err == nil {
		fmt.Printf("%v\n", value)
	}
}
```

Some formats (notably YAML) support a `Locator` interface for finding the line and column
for each data element, very useful for error messages:

```go
var yamlRepresentation = `
first:
  property1: Hello
  property2: 1.2
  property3: [ 1, 2 ]
`

func main() {
	if _, locator, err := ard.Decode(yamlRepresentation, "yaml", true); err == nil {
		if locator != nil {
			if line, column, ok := locator.Locate(
				ard.NewMapPathElement("first"),
				ard.NewFieldPathElement("property3"),
				ard.NewListPathElement(0),
			); ok {
				fmt.Printf("%d, %d\n", line, column) // 5, 16
			}
		}
	}
}
```

Unmarshal ARD into Go structs:

```go
var data = ard.Map{ // "ard.Map" is an alias for "map[any]any"
	"FirstName": "Gordon",
	"lastName":  "Freeman",
	"nicknames": ard.List{"Tigerbalm", "Stud Muffin"}, // "ard.List" is an alias for "[]any"
	"children": ard.List{
		ard.Map{
			"FirstName": "Bilbo",
		},
		ard.StringMap{ // "ard.StringMap" is an alias for "map[string]any"
			"FirstName": "Frodo",
		},
		nil,
	},
}

type Person struct {
	FirstName string    // property name will be used as field name
	LastName  string    `ard:"lastName"`   // "ard" tags work like familiar "json" and "yaml" tags
	Nicknames []string  `yaml:"nicknames"` // actually, go-ard will fall back to "yaml" tags by default
	Children  []*Person `json:"children"`  // ...and "json" tags, too
}

func main() {
	reflector := ard.NewReflector() // configurable; see documentation
	var p Person
	if err := reflector.Pack(data, &p); err == nil {
		fmt.Printf("%+v\n", p)
	}
}
```

Copy, merge, and compare:

```go
func main() {
	data_ := ard.SimpleCopy(data)
	fmt.Printf("%t\n", ard.Equals(data, data_))
	ard.MergeMaps(data, ard.Map{"role": "hero", "children": ard.List{"Gollum"}}, true)
	fmt.Printf("%v\n", data)
}
```

Node-based path traversal:

```go
var data = ard.Map{
	"first": ard.Map{
		"property1": "Hello",
		"property2": ard.StringMap{
			"second": ard.Map{
				"property1": 1}}}}

func main() {
	if p1, ok := ard.NewNode(data).Get("first", "property1").String(); ok {
		fmt.Println(p1)
	}
	if p2, ok := ard.NewNode(data).Get("first", "property2", "second", "property1").ConvertSimilar().Float(); ok {
		fmt.Printf("%f\n", p2)
	}
}
```

By default go-ard reads maps into `map[any]any`, but you can normalize for either `map[any]any` or
`map[string]map` (Go's JSON encoder *requires* the latter):

```go
import "encoding/json"

var data = ard.Map{ // remember, these are "map[any]any"
	"person": ard.Map{
		"age": uint(120),
	},
}

func main() {
	if data_, ok := ard.NormalizeStringMaps(data); ok { // otherwise JSON won't be able to encode the "map[any]any"
		json.NewEncoder(os.Stdout).Encode(data_)
	}
}
```

Introducing "cjson" (Compatible JSON) format that extends JSON with support for missing ARD
types: integers, unsigned integers, and maps with non-string keys:

```go
var data = ard.Map{
	"person": ard.Map{
		"age": uint(120),
	},
}

func main() {
	if data_, ok := ard.ToCompatibleJSON(data); ok { // will also normalize to "map[string]any"
		if j, err := json.Marshal(data_); err == nil {
			fmt.Println(string(j)) // {"map":{"age":{"$ard.uinteger":"120"}}}
			if data__, _, err := ard.Decode(string(j), "cjson", false); err == nil {
				fmt.Printf("%v\n", data__)
			}
		}
	}
}
```
