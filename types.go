package ard

import (
	"fmt"
	"time"

	"github.com/tliron/kutil/util"
)

//
// TypeName
//

type TypeName string

const (
	NoType TypeName = ""

	// Failsafe schema: https://yaml.org/spec/1.2/spec.html#id2802346
	TypeMap    TypeName = "ard.map"
	TypeList   TypeName = "ard.list"
	TypeString TypeName = "ard.string"

	// JSON schema: https://yaml.org/spec/1.2/spec.html#id2803231
	TypeBoolean TypeName = "ard.boolean"
	TypeInteger TypeName = "ard.integer"
	TypeFloat   TypeName = "ard.float"

	// Other schemas: https://yaml.org/spec/1.2/spec.html#id2805770
	TypeNull      TypeName = "ard.null"
	TypeBytes     TypeName = "ard.bytes"
	TypeTimestamp TypeName = "ard.timestamp"
)

// Returns a canonical name for all supported ARD types, including
// primitives, [Map], [List], and [time.Time]. Note that [StringMap]
// is not supported by this function.
//
// Unspported types will use [fmt.Sprintf]("%T").
func GetTypeName(value Value) TypeName {
	switch value.(type) {
	case Map:
		return TypeMap
	case List:
		return TypeList
	case string:
		return TypeString
	case bool:
		return TypeBoolean
	case int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint:
		return TypeInteger
	case float64, float32:
		return TypeFloat
	case nil:
		return TypeNull
	case []byte:
		return TypeBytes
	case time.Time:
		return TypeTimestamp
	default:
		return TypeName(fmt.Sprintf("%T", value))
	}
}

//
// TypeZeroes
//

var TypeZeroes = map[TypeName]Value{
	TypeMap:       make(Map), // YAML parser returns Map, but JSON parser returns StringMap
	TypeList:      List{},
	TypeString:    "",
	TypeBoolean:   false,
	TypeInteger:   int(0),       // YAML parser returns int
	TypeFloat:     float64(0.0), // YAML and JSON parsers return float64
	TypeNull:      nil,
	TypeBytes:     []byte{},
	TypeTimestamp: time.Time{}, // YAML parser returns time.Time
}

//
// TypeValidator
//

type TypeValidator = func(Value) bool

var TypeValidators = map[TypeName]TypeValidator{
	TypeMap:       IsMap,
	TypeList:      IsList,
	TypeString:    IsString,
	TypeBoolean:   IsBoolean,
	TypeInteger:   util.IsInteger,
	TypeFloat:     util.IsFloat,
	TypeNull:      IsNull,
	TypeBytes:     IsBytes,
	TypeTimestamp: IsTimestamp,
}

// Returns true if value is a [Map] (map[any]any).
//
// ([TypeValidator] signature)
func IsMap(value Value) bool {
	_, ok := value.(Map)
	return ok
}

// Returns true if value is a [List] ([]any).
//
// ([TypeValidator] signature)
func IsList(value Value) bool {
	_, ok := value.(List)
	return ok
}

// Returns true if value is a string.
//
// ([TypeValidator] signature)
func IsString(value Value) bool {
	_, ok := value.(string)
	return ok
}

// Returns true if value is a [bool].
//
// ([TypeValidator] signature)
func IsBoolean(value Value) bool {
	_, ok := value.(bool)
	return ok
}

// Returns true if value is nil.
//
// ([TypeValidator] signature)
func IsNull(value Value) bool {
	return value == nil
}

// Returns true if value is []byte.
//
// ([TypeValidator] signature)
func IsBytes(value Value) bool {
	_, ok := value.([]byte)
	return ok
}

// Returns true if value is a [time.Time].
//
// ([TypeValidator] signature)
func IsTimestamp(value Value) bool {
	_, ok := value.(time.Time)
	return ok
}

// Returns true if value is a string, bool, int64, int32, int16, int8, int,
// uint64, uint32, uint16, uint8, uint, float64, float32, nil, []byte, or
// [time.Time].
func IsPrimitiveType(value Value) bool {
	switch value.(type) {
	case string, bool, int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint, float64, float32, nil, []byte, time.Time:
		return true
	default:
		return false
	}
}
