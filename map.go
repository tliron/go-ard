package ard

import (
	"github.com/tliron/yamlkeys"
)

// Ensure data adheres to map[string]any
// (JSON encoding does not support map[any]any)
func EnsureStringMaps(stringMap StringMap) StringMap {
	stringMap_, _ := NormalizeStringMaps(stringMap)
	return stringMap_.(StringMap)
}

// Recursive
func StringMapToMap(stringMap StringMap) Map {
	map_ := make(Map)
	for key, value := range stringMap {
		map_[key], _ = NormalizeMaps(value)
	}
	return map_
}

// Recursive
func MapToStringMap(map_ Map) StringMap {
	stringMap := make(StringMap)
	for key, value := range map_ {
		stringMap[yamlkeys.KeyString(key)], _ = NormalizeStringMaps(value)
	}
	return stringMap
}
