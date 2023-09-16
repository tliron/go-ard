package ard

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/tliron/kutil/util"
)

/*
ARD-compatible extended JSON can encode features that exist in YAML but not normally
in JSON:

1) integers and unsigned integers are preserved as distinct from floats
2) raw bytes can be encoded using Base64
3) maps are allowed to have non-string keys

This particular implementation is not designed for performance but rather for
widest compability, relying on Go's built-in JSON support or 3rd-party
implementations compatible with it.

Inspired by: https://www.mongodb.com/docs/manual/reference/mongodb-extended-json/
*/

const (
	XJSONIntegerCode  = "$ard.integer"
	XJSONUIntegerCode = "$ard.uinteger"
	XJSONBytesCode    = "$ard.bytes"
	XJSONMapCode      = "$ard.map"
)

// Prepares an ARD [Value] for encoding via [xml.Encoder].
//
// If inPlace is false then the function is non-destructive:
// the returned data structure is a [ValidCopy] of the value
// argument. Otherwise, the value may be changed during
// preparation.
//
// The reflector argument can be nil, in which case a
// default reflector will be used.
func PrepareForEncodingXJSON(value Value, inPlace bool, reflector *Reflector) (any, error) {
	if !inPlace {
		var err error
		if value, err = ValidCopy(value, reflector); err != nil {
			return nil, err
		}
	}

	value, _ = PackXJSON(value)
	return value, nil
}

func PackXJSON(value Value) (any, bool) {
	switch value_ := value.(type) {
	case int:
		return XJSONInteger(int64(value_)), true
	case int64:
		return XJSONInteger(value_), true
	case int32:
		return XJSONInteger(int64(value_)), true
	case int16:
		return XJSONInteger(int64(value_)), true
	case int8:
		return XJSONInteger(int64(value_)), true

	case uint:
		return XJSONUInteger(uint64(value_)), true
	case uint64:
		return XJSONUInteger(value_), true
	case uint32:
		return XJSONUInteger(uint64(value_)), true
	case uint16:
		return XJSONUInteger(uint64(value_)), true
	case uint8:
		return XJSONUInteger(uint64(value_)), true

	case []byte:
		return XJSONBytes(value_), true

	case List:
		converted := false
		convertedList := make(List, len(value_))

		for index, element := range value_ {
			var converted_ bool
			element, converted_ = PackXJSON(element)
			convertedList[index] = element
			if converted_ {
				converted = true
			}
		}

		if converted {
			return convertedList, true
		}

	case StringMap:
		if escapedValue, ok := escapeXjsonStringMap(value_); ok {
			return escapedValue, true
		}

		convertedStringMap := make(StringMap)
		converted := false
		var converted_ bool

		for key, value__ := range value_ {
			if convertedStringMap[key], converted_ = PackXJSON(value__); converted_ {
				converted = true
			}
		}

		if converted {
			return convertedStringMap, true
		}

	case Map:
		if escapedValue, ok := escapeXjsonMap(value_); ok {
			return escapedValue, true
		}

		// We'll build two maps at the same time, but only return one
		convertedStringMap := make(StringMap)
		convertedXJsonMap := make(XJSONMap)
		useXJsonMap := false

		for key, value__ := range value_ {
			value__, _ = PackXJSON(value__)

			if key_, ok := key.(string); ok {
				convertedXJsonMap[key] = value__

				// We can stop building the stringMap if we switched to xJsonMap
				if !useXJsonMap {
					convertedStringMap[key_] = value__
				}
			} else {
				key, _ = PackXJSON(key)
				convertedXJsonMap[key] = value__
				useXJsonMap = true
			}
		}

		if useXJsonMap {
			return convertedXJsonMap, true
		} else {
			return convertedStringMap, true
		}
	}

	return value, false
}

func UnpackXJSON(value any, useStringMaps bool) (Value, bool) {
	switch value_ := value.(type) {
	case List:
		converted := false
		list := make(List, len(value_))

		for index, element := range value_ {
			var converted_ bool
			element, converted_ = UnpackXJSON(element, useStringMaps)
			list[index] = element
			if converted_ {
				converted = true
			}
		}

		if converted {
			return list, true
		}

	case Map:
		map_ := make(StringMap)
		for key, value__ := range value_ {
			map_[MapKeyToString(key)] = value__
		}
		return UnpackXJSON(map_, useStringMaps)

	case StringMap:
		if len(value_) == 1 {
			if integer, ok := UnpackXJSONInteger(value_); ok {
				return integer, true
			} else if uinteger, ok := UnpackXJSONUInteger(value_); ok {
				return uinteger, true
			} else if bytes, ok := UnpackXJSONBytes(value_); ok {
				return bytes, true
			} else if map_, ok := UnpackXJSONMap(value_, useStringMaps); ok {
				return map_, true
			} else {
				// Handle escape code:
				// $$ -> $
				for key, value__ := range value_ {
					if strings.HasPrefix(key, "$$") {
						key = key[1:]
						map_ := make(Map)
						value__, _ = UnpackXJSON(value__, useStringMaps)
						map_[key] = value__
						return map_, true
					}
				}
			}
		}

		if useStringMaps {
			map_ := make(StringMap)
			for key, value__ := range value_ {
				value__, _ = UnpackXJSON(value__, useStringMaps)
				map_[key] = value__
			}
			return map_, true
		} else {
			map_ := make(Map)
			for key, value__ := range value_ {
				value__, _ = UnpackXJSON(value__, useStringMaps)
				map_[key] = value__
			}
			return map_, true
		}
	}

	return value, false
}

//
// XJSONInteger
//

type XJSONInteger int64

// ([json.Marshaler] interface)
func (self XJSONInteger) MarshalJSON() ([]byte, error) {
	return json.Marshal(StringMap{
		XJSONIntegerCode: strconv.FormatInt(int64(self), 10),
	})
}

func UnpackXJSONInteger(code StringMap) (int64, bool) {
	if integer, ok := code[XJSONIntegerCode]; ok {
		if integer_, ok := integer.(string); ok {
			if integer__, err := strconv.ParseInt(integer_, 10, 64); err == nil {
				return integer__, true
			}
		}
	}
	return 0, false
}

//
// XJSONUInteger
//

type XJSONUInteger uint64

// ([json.Marshaler] interface)
func (self XJSONUInteger) MarshalJSON() ([]byte, error) {
	return json.Marshal(StringMap{
		XJSONUIntegerCode: strconv.FormatUint(uint64(self), 10),
	})
}

func UnpackXJSONUInteger(code StringMap) (uint64, bool) {
	if uinteger, ok := code[XJSONUIntegerCode]; ok {
		if uinteger_, ok := uinteger.(string); ok {
			if uinteger__, err := strconv.ParseUint(uinteger_, 10, 64); err == nil {
				return uinteger__, true
			}
		}
	}
	return 0, false
}

//
// XJSONBytes
//

type XJSONBytes []byte

// ([json.Marshaler] interface)
func (self XJSONBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(StringMap{
		XJSONBytesCode: util.ToBase64(self),
	})
}

func UnpackXJSONBytes(code StringMap) ([]byte, bool) {
	if bytes, ok := code[XJSONBytesCode]; ok {
		if bytes_, ok := bytes.(string); ok {
			if bytes__, err := util.FromBase64(bytes_); err == nil {
				return bytes__, true
			}
		}
	}
	return nil, false
}

//
// XJSONMap
//

type XJSONMap Map

// ([json.Marshaler] interface)
func (self XJSONMap) MarshalJSON() ([]byte, error) {
	list := make([]XJSONMapEntry, 0, len(self))
	for key, value := range self {
		list = append(list, XJSONMapEntry{key, value})
	}

	return json.Marshal(StringMap{
		XJSONMapCode: list,
	})
}

func UnpackXJSONMap(code StringMap, useStringMaps bool) (Map, bool) {
	if map_, ok := code[XJSONMapCode]; ok {
		if map__, ok := map_.(List); ok {
			map___ := make(Map)
			for _, entry := range map__ {
				if entry_, ok := UnpackXJSONMapEntry(entry, useStringMaps); ok {
					map___[entry_.Key] = entry_.Value
				} else {
					return nil, false
				}
			}
			return map___, true
		}
	}
	return nil, false
}

//
// XJSONMapEntry
//

type XJSONMapEntry struct {
	Key   Value `json:"key"`
	Value Value `json:"value"`
}

func UnpackXJSONMapEntry(entry Value, useStringMaps bool) (*XJSONMapEntry, bool) {
	if entry_, ok := entry.(StringMap); ok {
		if key, ok := entry_["key"]; ok {
			if value, ok := entry_["value"]; ok {
				key, _ = UnpackXJSON(key, useStringMaps)
				value, _ = UnpackXJSON(value, useStringMaps)
				return &XJSONMapEntry{
					Key:   key,
					Value: value,
				}, true
			}
		}
	}
	return nil, false
}

func escapeXjsonMap(map_ Map) (Value, bool) {
	if len(map_) == 1 {
		if value, ok := map_[XJSONIntegerCode]; ok {
			value, _ = PackXJSON(value)
			return StringMap{"$" + XJSONIntegerCode: value}, true
		} else if value, ok := map_[XJSONUIntegerCode]; ok {
			value, _ = PackXJSON(value)
			return StringMap{"$" + XJSONUIntegerCode: value}, true
		} else if value, ok := map_[XJSONBytesCode]; ok {
			value, _ = PackXJSON(value)
			return StringMap{"$" + XJSONBytesCode: value}, true
		} else if value, ok := map_[XJSONMapCode]; ok {
			value, _ = PackXJSON(value)
			return StringMap{"$" + XJSONMapCode: value}, true
		}
	}
	return nil, false
}

func escapeXjsonStringMap(map_ StringMap) (Value, bool) {
	if len(map_) == 1 {
		if value, ok := map_[XJSONIntegerCode]; ok {
			value, _ = PackXJSON(value)
			return StringMap{"$" + XJSONIntegerCode: value}, true
		} else if value, ok := map_[XJSONUIntegerCode]; ok {
			value, _ = PackXJSON(value)
			return StringMap{"$" + XJSONUIntegerCode: value}, true
		} else if value, ok := map_[XJSONBytesCode]; ok {
			value, _ = PackXJSON(value)
			return StringMap{"$" + XJSONBytesCode: value}, true
		} else if value, ok := map_[XJSONMapCode]; ok {
			value, _ = PackXJSON(value)
			return StringMap{"$" + XJSONMapCode: value}, true
		}
	}
	return nil, false
}
