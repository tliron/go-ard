package ard

import (
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

func NewNode(data any) *Node {
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

// Returns a copy of this node for which similar values are allows and converted
// to the requested type. For example, [Node.Float] on an integer or unsigned integer
// would convert it to float.
func (self *Node) ConvertSimilar() *Node {
	if self == NoNode {
		return NoNode
	}

	return &Node{self.Value, self.container, self.key, self.nilMeansZero, true}
}

// Returns ([string], true) if the node is [string].
//
// If [Node.ConvertSimilar] was called then will convert any value
// to a string representation and return true (unless we are [NoNode]).
// Values are converted using [ValueToString].
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

// Returns ([]byte, true) if the node is []byte.
//
// If [Node.ConvertSimilar] was called and the node is a string
// then will attempt to decode it as base64, with failures returning
// false, false.
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

// Returns ([int64], true) if the node is [int64], [int32],
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

// Returns ([uint64], true) if the node is [uint64], [uint32],
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

// Returns ([float64], true) if the node is [float64] or [float32].
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

// Returns ([bool], true) if the node is [bool].
//
// If [Node.ConvertSimilar] was called then will call [Node.String]
// and then [strconv.ParseBool], with failures returning (false, false).
// Thus "true", "1", and 1 (ints and floats) will all be interpreted as
// boolean true.
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

// Returns ([Map], true) if the node is [Map].
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

// Returns ([StringMap], true) if the node is [StringMap].
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

// Returns ([List], true) if the node is [List].
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

// Returns ([]string, true) if the node is [List] and all its
// elements are strings. Will optimize if the node is already a
// []string, which doesn't normally occur in ARD.
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

// Gets a [Map] node nested in other [Map] nodes by keys, recursively.
// Importantly, this function will create new [Map] nodes along the way
// if they don't already exist, so it can change the ARD structure.
//
// If a [Map] can't be created along the way (meaning the the key already
// exists and is not a [Map]), or if this node is itself not a [Map], will
// return false.
func (self *Node) NestedMap(keys ...Value) (Map, bool) {
	if self == NoNode {
		return nil, false
	}

	switch map_ := self.Value.(type) {
	case Map:
		for _, key := range keys {
			if value, ok := map_[key]; ok {
				// Try to use existing map
				if map_, ok = value.(Map); !ok {
					// It's not a map
					return nil, false
				}
			} else {
				// Create a new map
				map__ := make(Map)
				map_[key] = map__
				map_ = map__
			}
		}

		return map_, true

	default:
		return nil, false
	}
}

// Gets a [StringMap] node nested in other [StringMap] nodes by keys, recursively.
// Importantly, this function will create new [StringMap] nodes along the way
// if they don't already exist, so it can change the ARD structure.
//
// If a [StringMap] can't be created along the way (meaning the the key already
// exists and is not a [StringMap]), or if this node is itself not a [StringMap], will
// return false.
func (self *Node) NestedStringMap(keys ...Value) (StringMap, bool) {
	if self == NoNode {
		return nil, false
	}

	switch map_ := self.Value.(type) {
	case StringMap:
		for _, key := range keys {
			key_ := MapKeyToString(key)
			if value, ok := map_[key_]; ok {
				// Try to use existing map
				if map_, ok = value.(StringMap); !ok {
					// It's not a map
					return nil, false
				}
			} else {
				// Create a new map
				map__ := make(StringMap)
				map_[key_] = map__
				map_ = map__
			}
		}

		return map_, true

	default:
		return nil, false
	}
}

// Gets nested nodes from [Map] or [StringMap] by keys, recursively. Returns
// [NoNode] if a key is not found.
//
// For [StringMap] keys are converted using [MapKeyToString].
func (self *Node) Get(keys ...Value) *Node {
	if self == NoNode {
		return NoNode
	}

	self_ := self

	for _, key := range keys {
		switch map_ := self_.Value.(type) {
		case Map:
			if value, ok := map_[key]; ok {
				self_ = &Node{value, self_, key, self_.nilMeansZero, self_.convertSimilar}
			} else {
				return NoNode
			}

		case StringMap:
			if value, ok := map_[MapKeyToString(key)]; ok {
				self_ = &Node{value, self_, key, self_.nilMeansZero, self_.convertSimilar}
			} else {
				return NoNode
			}

		default:
			return NoNode
		}
	}

	return self_
}

// Puts values in [Map] or [StringMap] by key. Returns false if
// failed (not a [Map] or [StringMap]).
//
// For [StringMap] keys are converted using [MapKeyToString].
func (self *Node) Put(key Value, value Value) bool {
	if self == NoNode {
		return false
	}

	switch map_ := self.Value.(type) {
	case Map:
		map_[key] = value
		return true

	case StringMap:
		map_[MapKeyToString(key)] = value
		return true

	default:
		return false
	}
}

// For [StringMap] keys are converted using [MapKeyToString].
func (self *Node) PutNested(keys []Value, value Value) bool {
	if self == NoNode {
		return false
	}

	last := len(keys) - 1

	if last == -1 {
		return false
	}

	switch map_ := self.Value.(type) {
	case Map:
		if last > 0 {
			for _, key := range keys[:last] {
				if value_, ok := map_[key]; ok {
					// Try to use existing map
					if map_, ok = value_.(Map); !ok {
						return false
					}
				} else {
					// Create a new map
					map__ := make(Map)
					map_[key] = map__
					map_ = map__
				}
			}
		}

		map_[keys[last]] = value

		return true

	case StringMap:
		if last > 0 {
			for _, key := range keys[:last] {
				key_ := MapKeyToString(key)
				if value_, ok := map_[key_]; ok {
					// Try to use existing map
					if map_, ok = value_.(StringMap); !ok {
						return false
					}
				} else {
					// Create a new map
					map__ := make(StringMap)
					map_[key_] = map__
					map_ = map__
				}
			}
		}

		map_[MapKeyToString(keys[last])] = value

		return true

	default:
		return true
	}
}

// Appends a value to a [List]. Returns false if failed
// (not a [List]).
//
// Note that this function changes the container of this node, which
// is the actual owner of the slice.
func (self *Node) Append(value Value) bool {
	if self == NoNode {
		return false
	}

	if list, ok := self.Value.(List); ok {
		self.container.Put(self.key, append(list, value))
		return true
	} else {
		return false
	}
}

// Convenience function to convert a string path to keys
// usable for [Node.Get], [Node.PutNested], [Node.NestedMap],
// and [Node.NestedStringMap].
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
