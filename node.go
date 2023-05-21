package ard

import (
	"github.com/tliron/yamlkeys"
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

func (self *Node) NilMeansZero() *Node {
	if self != NoNode {
		return &Node{self.Value, self.container, self.key, true, self.convertSimilar}
	}
	return NoNode
}

func (self *Node) ConvertSimilar() *Node {
	if self != NoNode {
		return &Node{self.Value, self.container, self.key, self.nilMeansZero, true}
	}
	return NoNode
}

func (self *Node) Get(keys ...string) *Node {
	self_ := self
	if self_ != NoNode {
		for _, key := range keys {
			switch map_ := self_.Value.(type) {
			case StringMap:
				if value, ok := map_[key]; ok {
					self_ = &Node{value, self_, key, self_.nilMeansZero, self_.convertSimilar}
				} else {
					return NoNode
				}

			case Map:
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

func (self *Node) Append(value Value) bool {
	if self != NoNode {
		if list, ok := self.Value.(List); ok {
			self.container.Put(self.key, append(list, value))
			return true
		}
	}
	return false
}

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

func (self *Node) String() (string, bool) {
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return "", true
		}
		value, ok := self.Value.(string)
		return value, ok
	}
	return "", false
}

// Supports .ConvertSimilar()
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

// Supports .ConvertSimilar()
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

// Supports .ConvertSimilar()
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

func (self *Node) Boolean() (bool, bool) {
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return false, true
		} else if value, ok := self.Value.(bool); ok {
			return value, true
		}
	}
	return false, false
}

// Supports .ConvertSimilar()
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
					stringMap[yamlkeys.KeyString(key)] = value
				}
				return stringMap, true
			}
		}
	}
	return nil, false
}

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

func (self *Node) EnsureMap(keys ...string) (Map, bool) {
	if self != NoNode {
		if map_, ok := self.Value.(Map); ok {
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

func (self *Node) StringList() ([]string, bool) {
	if self != NoNode {
		if self.nilMeansZero && (self.Value == nil) {
			return nil, true
		}
		if value, ok := self.Value.(List); ok {
			list := make([]string, len(value))
			for index, element := range value {
				if list[index], ok = element.(string); !ok {
					return nil, false
				}
			}
			return list, true
		} else {
			return nil, false
		}
	}
	return nil, false
}
