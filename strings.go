package ard

import (
	"github.com/tliron/kutil/util"
	"github.com/tliron/yamlkeys"
)

// Provides consistent stringification of primitive ARD [Value].
//
// Non-ARD types will be converted via [util.ToString].
func ValueToString(value Value) string {
	return util.ToString(value)
}

// Provides consistent stringification of keys for ARD [StringMap].
//
// Used by functions such as [ConvertMapsToStringMaps] and
// [CopyMapsToStringMaps].
func MapKeyToString(key any) string {
	return yamlkeys.KeyString(key)
}
