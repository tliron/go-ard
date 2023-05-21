package ard

import (
	"fmt"
	"time"
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
	TypeInteger:   IsInteger,
	TypeFloat:     IsFloat,
	TypeNull:      IsNull,
	TypeBytes:     IsBytes,
	TypeTimestamp: IsTimestamp,
}

// Map = map[any]any
func IsMap(value Value) bool {
	_, ok := value.(Map)
	return ok
}

// List = []any
func IsList(value Value) bool {
	_, ok := value.(List)
	return ok
}

// string
func IsString(value Value) bool {
	_, ok := value.(string)
	return ok
}

// bool
func IsBoolean(value Value) bool {
	_, ok := value.(bool)
	return ok
}

// int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint
func IsInteger(value Value) bool {
	switch value.(type) {
	case int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint:
		return true
	}
	return false
}

// float64, float32
func IsFloat(value Value) bool {
	switch value.(type) {
	case float64, float32:
		return true
	}
	return false
}

func IsNull(value Value) bool {
	return value == nil
}

func IsBytes(value Value) bool {
	_, ok := value.([]byte)
	return ok
}

// time.Time
func IsTimestamp(value Value) bool {
	_, ok := value.(time.Time)
	return ok
}

func IsPrimitiveType(value Value) bool {
	switch value.(type) {
	case string, bool, int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint, float64, float32, nil, []byte, time.Time:
		return true
	default:
		return false
	}
}

// int64, int32, int16, int8, int
func ToInt64(value Value) int64 {
	switch value_ := value.(type) {
	case int64:
		return value_
	case int32:
		return int64(value_)
	case int16:
		return int64(value_)
	case int8:
		return int64(value_)
	case int:
		return int64(value_)
	case uint64:
		return int64(value_)
	case uint32:
		return int64(value_)
	case uint16:
		return int64(value_)
	case uint8:
		return int64(value_)
	case uint:
		return int64(value_)
	default:
		panic(fmt.Sprintf("not an integer: %T", value))
	}
}

// uint64, uint32, uint16, uint8, uint
func ToUInt64(value Value) uint64 {
	switch value_ := value.(type) {
	case uint64:
		return value_
	case uint32:
		return uint64(value_)
	case uint16:
		return uint64(value_)
	case uint8:
		return uint64(value_)
	case uint:
		return uint64(value_)
	case int64:
		return uint64(value_)
	case int32:
		return uint64(value_)
	case int16:
		return uint64(value_)
	case int8:
		return uint64(value_)
	case int:
		return uint64(value_)
	default:
		panic(fmt.Sprintf("not an integer: %T", value))
	}
}

// float64, float32
func ToFloat64(value Value) float64 {
	switch value_ := value.(type) {
	case float64:
		return value_
	case float32:
		return float64(value_)
	default:
		panic(fmt.Sprintf("not a float: %T", value))
	}
}
