package ard

import (
	"fmt"
	"regexp"
)

//
// PathElement
//

type PathElement struct {
	Type  PathElementType
	Value any // string for FieldPathType and MapPathType, int for ListPathType and SequencedListPathType
}

type PathElementType uint8

const (
	FieldPathType = iota
	MapPathType
	ListPathType
	SequencedListPathType
)

func NewFieldPathElement(name string) PathElement {
	return PathElement{FieldPathType, name}
}

func NewMapPathElement(name string) PathElement {
	return PathElement{MapPathType, name}
}

func NewListPathElement(index int) PathElement {
	return PathElement{ListPathType, index}
}

func NewSequencedListPathElement(index int) PathElement {
	return PathElement{SequencedListPathType, index}
}

//
// Path
//

type Path []PathElement

func (self Path) Append(element PathElement) Path {
	length := len(self)
	path := make(Path, length+1)
	copy(path, self)
	path[length] = element
	return path
}

func (self Path) AppendField(name string) Path {
	return self.Append(NewFieldPathElement(name))
}

func (self Path) AppendMap(name string) Path {
	return self.Append(NewMapPathElement(name))
}

func (self Path) AppendList(index int) Path {
	return self.Append(NewListPathElement(index))
}

func (self Path) AppendSequencedList(index int) Path {
	return self.Append(NewSequencedListPathElement(index))
}

var fieldPathElementEscapeRe = regexp.MustCompile(`([\".\[\]{}])`)

// fmt.Stringer interface
func (self Path) String() string {
	var path string

	for _, element := range self {
		switch element.Type {
		case FieldPathType:
			value := fieldPathElementEscapeRe.ReplaceAllString(element.Value.(string), "\\$1")
			if path == "" {
				path = value
			} else {
				path = fmt.Sprintf("%s.%s", path, value)
			}

		case MapPathType:
			value := element.Value.(string)
			path = fmt.Sprintf("%s[%q]", path, value)

		case ListPathType:
			value := element.Value.(int)
			path = fmt.Sprintf("%s[%d]", path, value)

		case SequencedListPathType:
			value := element.Value.(int)
			path = fmt.Sprintf("%s{%d}", path, value)
		}
	}

	return path
}
