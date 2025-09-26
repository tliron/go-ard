package ard

import (
	"fmt"
	"strconv"
	"time"

	"github.com/tliron/go-kutil/util"
	"gopkg.in/yaml.v3"
)

func ToYAMLDocumentNode(value Value, verbose bool, reflector *Reflector) (*yaml.Node, error) {
	if value_, err := ValidCopy(value, reflector); err == nil {
		if node, ok := ToYAMLNode(value_, verbose); ok {
			return &yaml.Node{
				Kind:    yaml.DocumentNode,
				Content: []*yaml.Node{node},
			}, nil
		} else {
			return nil, fmt.Errorf("unsupported value type: %T", value_)
		}
	} else {
		return nil, err
	}
}

func ToYAMLNode(value Value, verbose bool) (*yaml.Node, bool) {
	var node yaml.Node
	if verbose {
		node.Style = yaml.TaggedStyle
	}

	switch value_ := value.(type) {
	// Failsafe schema: https://yaml.org/spec/1.2/spec.html#id2802346

	case Map:
		node.Kind = yaml.MappingNode
		node.Tag = "!!map"
		node.Style = 0
		node.Content = make([]*yaml.Node, len(value_)*2)
		index := 0
		for k, v := range value_ {
			var ok bool
			if node.Content[index], ok = ToYAMLNode(k, verbose); ok {
				index += 1
				if node.Content[index], ok = ToYAMLNode(v, verbose); ok {
					index += 1
				} else {
					return nil, false
				}
			} else {
				return nil, false
			}
		}

	case StringMap:
		node.Kind = yaml.MappingNode
		node.Tag = "!!map"
		node.Style = 0
		node.Content = make([]*yaml.Node, len(value_)*2)
		index := 0
		for k, v := range value_ {
			var ok bool
			if node.Content[index], ok = ToYAMLNode(k, verbose); ok {
				index += 1
				if node.Content[index], ok = ToYAMLNode(v, verbose); ok {
					index += 1
				} else {
					return nil, false
				}
			} else {
				return nil, false
			}
		}

	case List:
		node.Kind = yaml.SequenceNode
		node.Tag = "!!seq"
		node.Style = 0
		node.Content = make([]*yaml.Node, len(value_))
		for index, v := range value_ {
			var ok bool
			if node.Content[index], ok = ToYAMLNode(v, verbose); !ok {
				return nil, false
			}
		}

	case string:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!str"
		node.Style |= yaml.DoubleQuotedStyle
		node.Value = value_

	// JSON schema: https://yaml.org/spec/1.2/spec.html#id2803231

	case nil:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!null"

	case bool:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!bool"
		node.Value = strconv.FormatBool(value_)

	case int64:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!int"
		node.Value = strconv.FormatInt(value_, 10)

	case int32:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!int"
		node.Value = strconv.FormatInt(int64(value_), 10)

	case int16:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!int"
		node.Value = strconv.FormatInt(int64(value_), 10)

	case int8:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!int"
		node.Value = strconv.FormatInt(int64(value_), 10)

	case int:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!int"
		node.Value = strconv.FormatInt(int64(value_), 10)

	case uint64:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!int"
		node.Value = strconv.FormatUint(value_, 10)

	case uint32:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!int"
		node.Value = strconv.FormatUint(uint64(value_), 10)

	case uint16:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!int"
		node.Value = strconv.FormatUint(uint64(value_), 10)

	case uint8:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!int"
		node.Value = strconv.FormatUint(uint64(value_), 10)

	case uint:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!int"
		node.Value = strconv.FormatUint(uint64(value_), 10)

	case float64:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!float"
		node.Value = fixFloat(strconv.FormatFloat(value_, 'g', -1, 64))

	case float32:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!float"
		node.Value = fixFloat(strconv.FormatFloat(float64(value_), 'g', -1, 32))

	// Other schemas: https://yaml.org/spec/1.2/spec.html#id2805770

	case []byte:
		// See: https://yaml.org/type/binary.html
		node.Kind = yaml.ScalarNode
		node.Tag = "!!binary"
		node.Value = util.ToBase64(value_)

	case time.Time:
		// See: https://yaml.org/type/timestamp.html
		node.Kind = yaml.ScalarNode
		node.Tag = "!!timestamp"
		node.Value = value_.Format(time.RFC3339Nano)

	default:
		return nil, false
	}

	return &node, true
}

func fixFloat(s string) string {
	// See: https://yaml.org/spec/1.2/spec.html#id2804092
	switch s {
	case "+Inf":
		return ".inf"
	case "-Inf":
		return "-.inf"
	case "NaN":
		return ".nan"
	}
	return s
}

func FindYAMLNode(node *yaml.Node, path ...PathElement) *yaml.Node {
	if len(path) == 0 {
		return node
	}

	switch node.Kind {
	case yaml.AliasNode:
		return FindYAMLNode(node.Alias, path...)

	case yaml.DocumentNode:
		if len(node.Content) > 0 {
			// Length *should* be 1
			return FindYAMLNode(node.Content[0], path...)
		}

	case yaml.MappingNode:
		pathElement := path[0]
		switch pathElement.Type {
		case FieldPathType, MapPathType:
			value := pathElement.Value.(string)

			// Content is a slice of pairs of key-followed-by-value
			length := len(node.Content)
			for i := 0; i < length; i += 2 {
				keyNode := node.Content[i]

				if i+1 >= length {
					// Length *should* be an even number
					return keyNode
				}

				// Is it in one of the merged values?
				if (keyNode.Kind == yaml.ScalarNode) && (keyNode.Tag == "!!merge") {
					valueNode := node.Content[i+1]
					foundNode := FindYAMLNode(valueNode, path...)
					if foundNode != valueNode {
						return foundNode
					}
				}

				// We only support comparisons with string keys
				if (keyNode.Kind == yaml.ScalarNode) && (keyNode.Tag == "!!str") && (keyNode.Value == value) {
					valueNode := node.Content[i+1]
					foundNode := FindYAMLNode(valueNode, path[1:]...)
					if foundNode == valueNode {
						// We will use the key node for the location instead of the value node
						return keyNode
					}
					return foundNode
				}
			}
		}

	case yaml.SequenceNode:
		pathElement := path[0]
		switch pathElement.Type {
		case ListPathType:
			index := pathElement.Value.(int)
			if index < len(node.Content) {
				return FindYAMLNode(node.Content[index], path[1:]...)
			}

		case SequencedListPathType:
			index := pathElement.Value.(int)
			if index < len(node.Content) {
				content := node.Content[index]
				if (content.Kind == yaml.MappingNode) && (len(content.Content) == 2) {
					// Content is a slice of pairs of key-followed-by-value
					return FindYAMLNode(content.Content[1], path[1:]...)
				} else {
					return content
				}
			}
		}
	}

	return node
}
