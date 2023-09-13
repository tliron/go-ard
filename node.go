package ard

import (
	"strconv"
)

//
// Node
//

type Node struct {
	Value Value

	container      *Node
	key            string
	nilMeansZero   bool
	convertSimilar bool
}

func NewNode(data any) *Node {
	return &Node{data, nil, "", false, false}
}

var NoNode = &Node{nil, nil, "", false, false}

// Returns a copy of this node for which nil values are allowed and interpreted as
// the zero value. For example, [Node.String] on nil would return an empty string.
func (self *Node) NilMeansZero() *Node {
	if self != NoNode {
		return &Node{self.Value, self.container, self.key, true, self.convertSimilar}
	}
	return NoNode
}

// Returns a copy of this node for which similar values are allows and converted
// to the requested type. For example, [Node.Float] on an integer or unsigned integer
// would convert it to float.
func (self *Node) ConvertSimilar() *Node {
	if self != NoNode {
		return &Node{self.Value, self.container, self.key, self.nilMeansZero, true}
	}
	return NoNode
}

// Gets nested nodes from [Map] or [StringMap] by keys, recursively. Returns
// [NoNode] if a key is not found.
func (self *Node) Get(keys ...string) *Node {
	self_ := self
	if self_ != NoNode {
		for _, key := range keys {
			switch map_ := self_.Value.(type) {
			case Map:
				if value, ok := map_[key]; ok {
					self_ = &Node{value, self_, key, self_.nilMeansZero, self_.convertSimilar}
				} else {
					return NoNode
				}

			case StringMap:
				if value, ok := map_[key]; ok {
					self_ = &Node{value, self_, key, self_.nilMeansZero, self_.convertSimilar}
				} else {
					return NoNode
				}

			default:
				return NoNode
			}
		}
	}
	return self_
}

// Puts values in [Map] or [StringMap] by key. Returns false if
// failed (not a [Map] or [StringMap]).
func (self *Node) Put(key string, value Value) bool {
	if self != NoNode {
		switch map_ := self.Value.(type) {
		case StringMap:
			map_[key] = value
			return true

		case Map:
			map_[key] = value
			return true
		}
	}
	return false
}

func (self *Node) PutNested(keys []string, value Value) bool {
	if self != NoNode {
		switch map_ := self.Value.(type) {
		case Map:
			last := len(keys) - 1

			if last == -1 {
				return false
			}

			if last > 0 {
				for _, lastPath := range keys[:last] {
					if value_, ok := map_[lastPath]; ok {
						if map_, ok = value_.(Map); !ok {
							return false
						}
					} else {
						map__ := make(Map)
						map_[lastPath] = map__
						map_ = map__
					}
				}
			}

			map_[keys[last]] = value

			return true

		case StringMap:
			last := len(keys) - 1

			if last == -1 {
				return false
			}

			if last > 0 {
				for _, lastPath := range keys[:last] {
					if value_, ok := map_[lastPath]; ok {
						if map_, ok = value_.(StringMap); !ok {
							return false
						}
					} else {
						map__ := make(StringMap)
						map_[lastPath] = map__
						map_ = map__
					}
				}
			}

			map_[keys[last]] = value

			return true
		}
	}

	return false
}

func (self *Node) EnsureMap(keys ...string) (Map, bool) {
	if self != NoNode {
		switch map_ := self.Value.(type) {
		case Map:
			for _, key := range keys {
				if value, ok := map_[key]; ok {
					if map__, ok := value.(Map); ok {
						map_ = map__
					} else {
						return nil, false
					}
				} else {
					map__ := make(Map)
					map_[key] = map__
					map_ = map__
				}
			}
			return map_, true
		}
	}
	return nil, false
}

// Appends a value to a [List]. Returns false if failed
// (not a [List]).
//
// Note that this function changes the container of this node, which
// is the one that actually holds the slice.
func (self *Node) Append(value Value) bool {
	if self != NoNode {
		if list, ok := self.Value.(List); ok {
			self.container.Put(self.key, append(list, value))
			return true
		}
	}
	return false
}

// Returns []byte, true if the node is []byte.
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as an empty []byte.
func (self *Node) Bytes() ([]byte, bool) {
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return nil, true
		}
		value, ok := self.Value.([]byte)
		return value, ok
	}
	return nil, false
}

// Returns [string], true if the node is [string].
//
// If [Node.ConvertSimilar] was called then will convert any value
// to a string representation and return true (unless we are [NoNode]).
// Values are converted using [MapKeyToString].
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as an empty string.
func (self *Node) String() (string, bool) {
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return "", true
		}
		value, ok := self.Value.(string)
		if !ok && self.convertSimilar {
			return MapKeyToString(value), true
		}
		return value, ok
	}
	return "", false
}

// Returns [int64], true if the node is [int64], [int32],
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
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return 0, true
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
		}
		if self.convertSimilar {
			switch value := self.Value.(type) {
			case uint64:
				return int64(value), true
			case uint32:
				return int64(value), true
			case uint16:
				return int64(value), true
			case uint8:
				return int64(value), true
			case uint:
				return int64(value), true
			case float64:
				return int64(value), true
			case float32:
				return int64(value), true
			}
		}
	}
	return 0, false
}

// Returns [uint64], true if the node is [uint64], [uint32],
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
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return 0, true
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
		}
		if self.convertSimilar {
			switch value := self.Value.(type) {
			case int64:
				return uint64(value), true
			case int32:
				return uint64(value), true
			case int16:
				return uint64(value), true
			case int8:
				return uint64(value), true
			case int:
				return uint64(value), true
			case float64:
				return uint64(value), true
			case float32:
				return uint64(value), true
			}
		}
	}
	return 0, false
}

// Returns [float64], true if the node is [float64] or [float32].
//
// If [Node.ConvertSimilar] was called then will convert [int64],
// [int32], [int16], [int8], [int], [uint64], [uint32], [uint16],
// [uint8], [uint] to a [float64] and return true (unless we are [NoNode]).
// Precision may be lost.
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as 0.
func (self *Node) Float() (float64, bool) {
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return 0.0, true
		}
		switch value := self.Value.(type) {
		case float64:
			return value, true
		case float32:
			return float64(value), true
		}
		if self.convertSimilar {
			switch value := self.Value.(type) {
			case int64:
				return float64(value), true
			case int32:
				return float64(value), true
			case int16:
				return float64(value), true
			case int8:
				return float64(value), true
			case int:
				return float64(value), true
			case uint64:
				return float64(value), true
			case uint32:
				return float64(value), true
			case uint16:
				return float64(value), true
			case uint8:
				return float64(value), true
			case uint:
				return float64(value), true
			}

		}
	}
	return 0.0, false
}

// Returns [bool], true if the node is [bool].
//
// If [Node.ConvertSimilar] was called then will call [Node.String]
// and then [strconv.ParseBool], with failures returning false, false.
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as false.
func (self *Node) Boolean() (bool, bool) {
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return false, true
		}
		value, ok := self.Value.(bool)
		if !ok && self.convertSimilar {
			if string_, ok := self.String(); ok {
				if value_, err := strconv.ParseBool(string_); err == nil {
					return value_, true
				} else {
					return false, false
				}
			} else {
				return false, false
			}
		}
		return value, ok
	}
	return false, false
}

// Returns [Map], true if the node is [Map].
//
// If [Node.ConvertSimilar] was called then will convert [StringMap]
// to a [Map] and return true.
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as an empty [Map].
func (self *Node) Map() (Map, bool) {
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return make(Map), true
		} else if map_, ok := self.Value.(Map); ok {
			return map_, true
		} else if self.convertSimilar {
			if stringMap, ok := self.Value.(StringMap); ok {
				map_ := make(Map)
				for key, value := range stringMap {
					map_[key] = value
				}
				return map_, true
			}
		}
	}
	return nil, false
}

// Returns [StringMap], true if the node is [StringMap].
//
// If [Node.ConvertSimilar] was called then will convert [Map]
// to a [StringMap] and return true. Keys are converted using
// [MapKeyToString].
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as an empty [StringMap].
func (self *Node) StringMap() (StringMap, bool) {
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return make(StringMap), true
		} else if stringMap, ok := self.Value.(StringMap); ok {
			return stringMap, true
		} else if self.convertSimilar {
			if map_, ok := self.Value.(Map); ok {
				stringMap := make(StringMap)
				for key, value := range map_ {
					stringMap[MapKeyToString(key)] = value
				}
				return stringMap, true
			}
		}
	}
	return nil, false
}

// Returns [List], true if the node is [List].
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as an empty [List].
func (self *Node) List() (List, bool) {
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return nil, true
		}
		list, ok := self.Value.(List)
		return list, ok
	}
	return nil, false
}

// Returns []string, true if the node is [List] and all its
// elements are strings.
//
// If [Node.ConvertSimilar] was called then will convert all
// elements to their string representations and return true.
// Values are converted using [MapKeyToString].
//
// By default will fail on nil values. Call [Node.NilMeansZero]
// to interpret nil as an empty [List].
func (self *Node) StringList() ([]string, bool) {
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return nil, true
		}
		if value, ok := self.Value.(List); ok {
			list := make([]string, len(value))
			for index, element := range value {
				string_, ok := element.(string)
				if !ok && self.convertSimilar {
					string_ = MapKeyToString(element)
				} else {
					return nil, false
				}
				list[index] = string_
			}
			return list, true
		} else {
			return nil, false
		}
	}
	return nil, false
}
