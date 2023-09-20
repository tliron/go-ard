package ard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tliron/kutil/util"
)

//
// Node
//

type Node struct {
	Value Value

	container      *Node
	key            Value
	nilMeansZero   bool
	convertSimilar bool
}

// Creates an extractable, convertible, traversable, and modifiable wrapper
// (a [Node]) for an ARD [Value].
func With(data any) *Node {
	return &Node{data, nil, "", false, false}
}

// This singleton is returned from all node functions when
// no node is found.
var NoNode = &Node{nil, nil, "", false, false}

// Returns a copy of this node for which nil values are allowed and interpreted as
// the zero value. For example, [Node.String] on nil would return an empty string.
func (self *Node) NilMeansZero() *Node {
	if self == NoNode {
		return NoNode
	}

	return &Node{self.Value, self.container, self.key, true, self.convertSimilar}
}

// Returns a copy of this node for which similarly-typed values are allowed and
// converted to the requested type. For example, [Node.Float] on an integer or
// unsigned integer would convert it to float.
func (self *Node) ConvertSimilar() *Node {
	if self == NoNode {
		return NoNode
	}

	return &Node{self.Value, self.container, self.key, self.nilMeansZero, true}
}

// Returns ([string], true) if the node is a [string].
//
// If [Node.ConvertSimilar] was called then will convert any value
// to a string representation using [ValueToString] and return true
// (unless we are [NoNode]).
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as an empty string.
func (self *Node) String() (string, bool) {
	if self == NoNode {
		return "", false
	}

	switch value := self.Value.(type) {
	case string:
		return value, true

	case nil:
		if self.nilMeansZero {
			return "", true
		}

	default:
		if self.convertSimilar {
			return ValueToString(value), true
		}
	}

	return "", false
}

// Returns ([]byte, true) if the node is a []byte.
//
// If [Node.ConvertSimilar] was called and the node is a [string]
// then will attempt to decode it as base64, with failures returning
// (false, false).
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as an empty []byte.
func (self *Node) Bytes() ([]byte, bool) {
	if self == NoNode {
		return nil, false
	}

	switch value := self.Value.(type) {
	case []byte:
		return value, true

	case nil:
		if self.nilMeansZero {
			return []byte{}, true
		}

	default:
		if self.convertSimilar {
			if string_, ok := self.Value.(string); ok {
				if value_, err := util.FromBase64(string_); err == nil {
					return value_, true
				}
			}
		}
	}

	return nil, false
}

// Returns ([int64], true) if the node is an [int64], [int32],
// [int16], [int8], or [int].
//
// If [Node.ConvertSimilar] was called then will convert [uint64],
// [uint32], [uint16], [uint8], [uint], [float64], and [float32]
// to an [int64] and return true (unless we are [NoNode]). Precision
// may be lost.
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as 0.
func (self *Node) Integer() (int64, bool) {
	if self == NoNode {
		return 0, false
	}

	switch value := self.Value.(type) {
	case int64:
		return value, true
	case int32:
		return int64(value), true
	case int16:
		return int64(value), true
	case int8:
		return int64(value), true
	case int:
		return int64(value), true

	case nil:
		if self.nilMeansZero {
			return 0, true
		}

	default:
		if self.convertSimilar {
			return util.ToInt64(self.Value)
		}
	}

	return 0, false
}

// Returns ([uint64], true) if the node is an [uint64], [uint32],
// [uint16], [uint8], or [uint].
//
// If [Node.ConvertSimilar] was called then will convert [int64],
// [int32], [int16], [int8], [int], [float64], and [float32]
// to an [uint64] and return true (unless we are [NoNode]). Precision
// may be lost.
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as 0.
func (self *Node) UnsignedInteger() (uint64, bool) {
	if self == NoNode {
		return 0, false
	}

	switch value := self.Value.(type) {
	case uint64:
		return value, true
	case uint32:
		return uint64(value), true
	case uint16:
		return uint64(value), true
	case uint8:
		return uint64(value), true
	case uint:
		return uint64(value), true

	case nil:
		if self.nilMeansZero {
			return 0, true
		}

	default:
		if self.convertSimilar {
			return util.ToUInt64(self.Value)
		}
	}

	return 0, false
}

// Returns ([float64], true) if the node is a [float64] or a [float32].
//
// If [Node.ConvertSimilar] was called then will convert [int64],
// [int32], [int16], [int8], [int], [uint64], [uint32], [uint16],
// [uint8], [uint] to a [float64] and return true (unless we are [NoNode]).
// Precision may be lost.
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as 0.0.
func (self *Node) Float() (float64, bool) {
	if self == NoNode {
		return 0.0, false
	}

	switch value := self.Value.(type) {
	case float64:
		return value, true
	case float32:
		return float64(value), true

	case nil:
		if self.nilMeansZero {
			return 0.0, true
		}

	default:
		if self.convertSimilar {
			return util.ToFloat64(self.Value)
		}
	}

	return 0.0, false
}

// Returns ([bool], true) if the node is a [bool].
//
// If [Node.ConvertSimilar] was called then will call [Node.String]
// and then [strconv.ParseBool], with failures returning (false, false).
// Thus "true", "1", and numerical 1 (both ints and floats) will all be
// interpreted as boolean true.
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as false.
func (self *Node) Boolean() (bool, bool) {
	if self == NoNode {
		return false, false
	}

	switch value := self.Value.(type) {
	case bool:
		return value, true

	case nil:
		if self.nilMeansZero {
			return false, true
		}

	default:
		if self.convertSimilar {
			if string_, ok := self.String(); ok {
				if value_, err := strconv.ParseBool(string_); err == nil {
					return value_, true
				}
			}
		}
	}

	return false, false
}

// Returns ([Map], true) if the node is a [Map].
//
// If [Node.ConvertSimilar] was called then will convert [StringMap]
// to a [Map] and return true.
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as an empty [Map].
func (self *Node) Map() (Map, bool) {
	if self == NoNode {
		return nil, false
	}

	switch value := self.Value.(type) {
	case Map:
		return value, true

	case nil:
		if self.nilMeansZero {
			return make(Map), true
		}

	case StringMap:
		if self.convertSimilar {
			map_ := make(Map)

			for key, value_ := range value {
				map_[key] = value_
			}

			return map_, true
		}
	}

	return nil, false
}

// Returns ([StringMap], true) if the node is a [StringMap].
//
// If [Node.ConvertSimilar] was called then will convert [Map]
// to a [StringMap] and return true. Keys are converted using
// [MapKeyToString].
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as an empty [StringMap].
func (self *Node) StringMap() (StringMap, bool) {
	if self == NoNode {
		return nil, false
	}

	switch value := self.Value.(type) {
	case StringMap:
		return value, true

	case nil:
		if self.nilMeansZero {
			return make(StringMap), true
		}

	case Map:
		if self.convertSimilar {
			stringMap := make(StringMap)

			for key, value_ := range value {
				stringMap[MapKeyToString(key)] = value_
			}

			return stringMap, true
		}
	}

	return nil, false
}

// Returns ([List], true) if the node is a [List].
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as an empty [List].
func (self *Node) List() (List, bool) {
	if self == NoNode {
		return nil, false
	}

	switch value := self.Value.(type) {
	case List:
		return value, true

	case nil:
		if self.nilMeansZero {
			return List{}, true
		}
	}

	return nil, false
}

// Returns a new ([]string, true) if the node is [List] and all
// its elements are strings. Will optimize and avoid copying if the
// node is already a []string, which doesn't normally occur in ARD.
//
// If [Node.ConvertSimilar] was called then will convert all
// [List] elements to their string representations and return true.
// Values are converted using [ValueToString].
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as an empty []string.
func (self *Node) StringList() ([]string, bool) {
	if self == NoNode {
		return nil, false
	}

	switch value := self.Value.(type) {
	case []string:
		return value, true

	case List:
		list := make([]string, len(value))
		for index, element := range value {
			string_, ok := element.(string)

			if !ok {
				if self.convertSimilar {
					string_ = ValueToString(element)
				} else {
					return nil, false
				}
			}

			list[index] = string_
		}

		return list, true

	case nil:
		if self.nilMeansZero {
			return []string{}, true
		}
	}

	return nil, false
}

// Sets the value of this node and its key in the containing map.
//
// Will fail and return false if there is no containing node or it's
// not [Map] or [StringMap].
func (self *Node) Set(value Value) bool {
	if self == NoNode {
		return false
	}

	if self.container != nil {
		switch self.container.Value.(type) {
		case Map, StringMap:
			putInMap(self.container.Value, self.key, value)
			self.Value = value
			return true
		}
	}

	return false
}

// Appends a value to a [List] and calls [Node.Set].
//
// Will fail and return false if there is no containing node or
// it's not [Map] or [StringMap], or if this node is not a [List].
func (self *Node) Append(value Value) bool {
	if self == NoNode {
		return false
	}

	if list, ok := self.Value.(List); ok {
		return self.Set(append(list, value))
	}

	return false
}

// Deletes this node's key from the containing node's map.
//
// Will fail and return false if there is no containing node or
// it's not [Map] or [StringMap].
func (self *Node) Delete() bool {
	if self == NoNode {
		return false
	}

	if self.container != nil {
		switch self.container.Value.(type) {
		case Map, StringMap:
			deleteFromMap(self.container.Value, self.key)
			self.container = nil
			self.key = nil
			self.Value = nil
			return true
		}
	}

	return false
}

// Gets a node from a nested [Map] or [StringMap] by recursively following keys.
// Returns [NoNode] if any key is not found.
//
// For [StringMap] keys are converted using [MapKeyToString].
func (self *Node) Get(keys ...Value) *Node {
	return self.get(keys, false)
}

// Gets a node from a nested [Map] or [StringMap] by recursively following keys.
// Along the way, if a nested node does not exist then a map will be created and
// put in its containing node. The type of the created map will match that of the
// containing map, either [Map] or [StringMap]. Returns [NoNode] if a nested node
// already exists and is not a [Map] or [StringMap]. Otherwise the returned node
// will have a nil value if the last key does not exist in its container. Thus
// note that if you called [Node.NilMeansZero] you will still get a value even
// though the last key does not exist in its container.
//
// For [StringMap] keys are converted using [MapKeyToString].
func (self *Node) ForceGet(keys ...Value) *Node {
	return self.get(keys, true)
}

// Convenience method to call [Node.Get] with [PathToKeys].
func (self *Node) GetPath(path string, separator string) *Node {
	return self.Get(PathToKeys(path, separator)...)
}

// Convenience method to call [Node.ForceGet] with [PathToKeys].
func (self *Node) ForceGetPath(path string, separator string) *Node {
	return self.ForceGet(PathToKeys(path, separator)...)
}

// Convenience function to convert a string path to keys
// usable for [Node.Get] and [Node.ForceGet].
//
// Does a [strings.Split] with the provided separator.
func PathToKeys(path string, separator string) []Value {
	keys := strings.Split(path, separator)
	keys_ := make([]Value, len(keys))
	for index, key := range keys {
		keys_[index] = key
	}
	return keys_
}

// Utils

func (self *Node) get(keys []Value, force bool) *Node {
	if self == NoNode {
		return NoNode
	}

	last := len(keys) - 1

	if last == -1 {
		return NoNode
	}

	switch self.Value.(type) {
	case Map, StringMap:
		current := self

		// Iterate all keys except last (all excpeted to be maps)
		if last > 0 {
			for _, key := range keys[:last] {
				// Try to use existing map
				if map_, ok, isMap := getFromMap(current.Value, key); ok {
					if isMap {
						// Key exists and is a map
						current = &Node{map_, current, key, current.nilMeansZero, current.convertSimilar}
					} else {
						// Key exists but is not a map
						return NoNode
					}
				} else if force {
					// Create a new map (same type as current)
					var childMap any

					switch currentMap := current.Value.(type) {
					case Map:
						childMap = make(Map)
						currentMap[key] = childMap

					case StringMap:
						childMap = make(StringMap)
						currentMap[MapKeyToString(key)] = childMap

					default:
						panic(fmt.Sprintf("not a map: %T", current))
					}

					current = &Node{childMap, current, key, current.nilMeansZero, current.convertSimilar}
				} else {
					return NoNode
				}
			}
		}

		// Last key
		lastKey := keys[last]
		if value, ok, _ := getFromMap(current.Value, lastKey); ok {
			return &Node{value, current, lastKey, current.nilMeansZero, current.convertSimilar}
		} else if force {
			return &Node{nil, current, lastKey, current.nilMeansZero, current.convertSimilar}
		}
	}

	return NoNode
}

// value, exists, isMap
func getFromMap(value any, key Value) (any, bool, bool) {
	switch map_ := value.(type) {
	case Map:
		if value_, ok := map_[key]; ok {
			switch value_.(type) {
			case Map, StringMap:
				return value_, true, true
			default:
				return value_, true, false
			}
		}

	case StringMap:
		if value_, ok := map_[MapKeyToString(key)]; ok {
			switch value_.(type) {
			case Map, StringMap:
				return value_, true, true
			default:
				return value_, true, false
			}
		}
	}

	return nil, false, false
}

func putInMap(map_ any, key Value, value Value) {
	switch map__ := map_.(type) {
	case Map:
		map__[key] = value

	case StringMap:
		map__[MapKeyToString(key)] = value

	default:
		panic(fmt.Sprintf("not a map: %T", map_))
	}
}

func deleteFromMap(map_ any, key Value) {
	switch map__ := map_.(type) {
	case Map:
		delete(map__, key)

	case StringMap:
		delete(map__, MapKeyToString(key))

	default:
		panic(fmt.Sprintf("not a map: %T", map_))
	}
}
