package ard

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/tliron/kutil/reflection"
	"github.com/tliron/kutil/util"
)

type StructFieldNameMapperFunc func(fieldName string) string

// Structs can implement their own custom packing with this interface.
type FromARD interface {
	FromARD(reflector *Reflector) (any, error)
}

// Structs can implement their own custom unpacking with this interface.
type ToARD interface {
	ToARD(reflector *Reflector) (any, error)
}

//
// Reflector
//

var defaultStructFieldTags = []string{"ard", "yaml", "json"}

type Reflector struct {
	// When true, non-existing struct fields will be ignored when packing.
	// Otherwise, will result in a packing error.
	IgnoreMissingStructFields bool

	// When true, nil values will be packed into the zero value for the target type.
	// Otherwise, only nullable types will be supported in the target: pointers, maps,
	// and slices, and other types will result in a packing error.
	NilMeansZero bool

	// Used for both packing and unpacking. Processed in order.
	StructFieldTags []string

	// While StructFieldTags can be used to specify specific unpacked names, when
	// untagged this function, if set, will be used for translating field names to
	// their unpacked names.
	StructFieldNameMapper StructFieldNameMapperFunc

	reflectFieldsCache sync.Map
}

// Creates a reflector with default struct field tags:
// "ard", "yaml", "json".
func NewReflector() *Reflector {
	return &Reflector{StructFieldTags: defaultStructFieldTags}
}

// Packs an ARD value into Go types, recursively.
//
// For Go struct field names, keys are converted from [Map] using
// [MapKeyToString].
//
// Structs can provide their own custom packing by implementing the
// [FromARD] interface.
//
// packedValuePtr must be a pointer.
func (self *Reflector) Pack(value Value, packedValuePtr any) error {
	packedValuePtr_ := reflect.ValueOf(packedValuePtr)
	if packedValuePtr_.Kind() == reflect.Pointer {
		return self.pack(nil, value, packedValuePtr_)
	} else {
		return fmt.Errorf("target is not a pointer: %T", packedValuePtr)
	}
}

// Unpacks Go types to ARD, recursively. [Map] is used for Go structs
// and maps.
//
// For Go struct field names, keys are converted to [StringMap]
// using [MapKeyToString].
//
// Structs can provide their own custom unpacking by implementing the
// [ToARD] interface.
//
// packedValuePtr must be a pointer.
func (self *Reflector) Unpack(packedValue any) (Value, error) {
	return self.unpack(nil, reflect.ValueOf(packedValue), false)
}

// Unpacks Go types to ARD, recursively. [StringMap] is used for Go
// structs and maps.
//
// Keys are converted using [MapKeyToString].
//
// Structs can provide their own custom unpacking by implementing the
// [ToARD] interface.
//
// packedValue can be a value or a pointer.
func (self *Reflector) UnpackStringMaps(packedValue any) (Value, error) {
	return self.unpack(nil, reflect.ValueOf(packedValue), true)
}

func (self *Reflector) pack(path Path, value Value, packedValue reflect.Value) error {
	packedType := packedValue.Type()

	// Dereference pointers
	for packedType.Kind() == reflect.Pointer {
		if value == nil {
			packedValue.SetZero()
			return nil
		}

		packedType = packedType.Elem()

		if packedValue.IsNil() {
			// Point to a new value
			packedValue.Set(reflect.New(packedType))
		}

		packedValue = packedValue.Elem()
	}

	switch value_ := value.(type) {
	case nil:
		if self.NilMeansZero {
			packedValue.SetZero()
		} else {
			kind := packedValue.Kind()
			if (kind != reflect.Map) && (kind != reflect.Slice) {
				return fmt.Errorf("%s is not a pointer, map, or slice: %s", path.String(), packedType.String())
			}
		}

	case string:
		if packedValue.Kind() == reflect.String {
			packedValue.SetString(value_)
		} else {
			return fmt.Errorf("%s is not a string: %s", path.String(), packedType.String())
		}

	case bool:
		if packedValue.Kind() == reflect.Bool {
			packedValue.SetBool(value_)
		} else {
			return fmt.Errorf("%s is not a bool: %s", path.String(), packedType.String())
		}

	case int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint, float64, float32:
		if reflection.IsInteger(packedValue.Kind()) {
			value__, _ := util.ToInt64(value_)
			packedValue.SetInt(value__)
		} else if reflection.IsUInteger(packedValue.Kind()) {
			value__, _ := util.ToUInt64(value_)
			packedValue.SetUint(value__)
		} else if reflection.IsFloat(packedValue.Kind()) {
			value__, _ := util.ToFloat64(value_)
			packedValue.SetFloat(value__)
		} else {
			return fmt.Errorf("%s is not a number: %s", path.String(), packedType.String())
		}

	case []byte, time.Time: // as-is values
		if packedType == reflect.TypeOf(value_) {
			packedValue.Set(reflect.ValueOf(value_))
		} else {
			return fmt.Errorf("%s is not a %T: %s", path.String(), value, packedType.String())
		}

	case List:
		if packedValue.Kind() == reflect.Slice {
			elemType := packedType.Elem()
			length := len(value_)
			list := reflect.MakeSlice(reflect.SliceOf(elemType), length, length)
			for index, elem := range value_ {
				if err := self.pack(path.AppendList(index), elem, list.Index(index)); err != nil {
					return err
				}
			}
			packedValue.Set(list)
		} else {
			return fmt.Errorf("%s is not a slice: %s", path.String(), packedType.String())
		}

	case Map:
		switch packedValue.Kind() {
		case reflect.Map:
			if packedValue.IsNil() {
				packedValue.Set(reflect.MakeMap(packedType))
			}

			keyType := packedType.Key()
			valueType := packedType.Elem()
			for k, v := range value_ {
				k_ := reflect.New(keyType)
				path_ := path.AppendMap(MapKeyToString(k_))
				if err := self.pack(path_, k, k_); err == nil {
					v_ := reflect.New(valueType)
					if err := self.pack(path_, v, v_); err == nil {
						packedValue.SetMapIndex(k_.Elem(), v_.Elem())
					} else {
						return fmt.Errorf("map value for %s", err.Error())
					}
				} else {
					return fmt.Errorf("map key for %s", err.Error())
				}
			}

		case reflect.Struct:
			// Support FromARD interface
			if fromArd, ok := packedValue.Interface().(FromARD); ok {
				if value__, err := fromArd.FromARD(self); err == nil {
					packedValue.Set(reflect.ValueOf(value__))
					return nil
				} else {
					return err
				}
			}

			reflectFields := self.newReflectFields(packedType)
			for k, v := range value_ {
				if err := self.packStructField(path, packedValue, MapKeyToString(k), v, reflectFields); err != nil {
					return err
				}
			}

		default:
			return fmt.Errorf("%s is not a map or struct: %s", path.String(), packedType.String())
		}

	case StringMap:
		switch packedValue.Kind() {
		case reflect.Map:
			if packedValue.IsNil() {
				packedValue.Set(reflect.MakeMap(packedType))
			}

			keyType := packedType.Key()
			valueType := packedType.Elem()
			for k, v := range value_ {
				k_ := reflect.New(keyType)
				path_ := path.AppendMap(MapKeyToString(k_))
				if err := self.pack(path_, k, k_); err == nil {
					v_ := reflect.New(valueType)
					if err := self.pack(path_, v, v_); err == nil {
						packedValue.SetMapIndex(k_.Elem(), v_.Elem())
					} else {
						return fmt.Errorf("map value for %s", err.Error())
					}
				} else {
					return fmt.Errorf("map key for %s", err.Error())
				}
			}

		case reflect.Struct:
			if fromArd, ok := packedValue.Interface().(FromARD); ok {
				if value__, err := fromArd.FromARD(self); err == nil {
					packedValue.Set(reflect.ValueOf(value__))
					return nil
				} else {
					return err
				}
			}

			reflectFields := self.newReflectFields(packedType)
			for k, v := range value_ {
				if err := self.packStructField(path, packedValue, k, v, reflectFields); err != nil {
					return err
				}
			}

		default:
			return fmt.Errorf("%s is not a map or struct: %s", path.String(), packedType.String())
		}

	default:
		return fmt.Errorf("%s is of unsupported type: %s", path.String(), packedType.String())
	}

	return nil
}

func (self *Reflector) packStructField(structPath Path, structValue reflect.Value, fieldName string, value Value, fieldNames reflectFields) error {
	path := structPath.AppendField(fieldName)
	field := fieldNames.getField(structValue, fieldName)

	if !field.IsValid() {
		if self.IgnoreMissingStructFields {
			return nil
		} else {
			return fmt.Errorf("%s does not exist", path.String())
		}
	}

	if !field.CanSet() {
		return fmt.Errorf("%s cannot be set", path.String())
	}

	return self.pack(path, value, field)
}

var time_ time.Time
var timeType = reflect.TypeOf(time_)

func (self *Reflector) unpack(path Path, packedValue reflect.Value, useStringMaps bool) (Value, error) {
	// Support ToARD interface
	if toArd, ok := packedValue.Interface().(ToARD); ok {
		return toArd.ToARD(self)
	}

	packedType := packedValue.Type()
	kind := packedType.Kind()

	// Dereference pointers
	for (kind == reflect.Pointer) || (kind == reflect.Interface) {
		if packedValue.IsNil() {
			return nil, nil
		}

		packedValue = packedValue.Elem()

		if toArd, ok := packedValue.Interface().(ToARD); ok {
			return toArd.ToARD(self)
		}

		packedType = packedValue.Type()
		kind = packedType.Kind()
	}

	if packedType == timeType {
		return packedValue.Interface(), nil
	}

	switch kind {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Float64, reflect.Float32:
		return packedValue.Interface(), nil

	case reflect.Slice:
		if packedType.Elem().Kind() == reflect.Uint8 {
			// []byte
			return packedValue.Interface(), nil
		}

		length := packedValue.Len()
		list := make(List, length)
		for index := 0; index < length; index++ {
			value_ := packedValue.Index(index)
			var err error
			if list[index], err = self.unpack(path.AppendList(index), value_, useStringMaps); err != nil {
				return nil, err
			}
		}
		return list, nil

	case reflect.Map:
		if useStringMaps {
			map_ := make(StringMap)
			keys := packedValue.MapKeys()
			for _, key := range keys {
				path_ := path.AppendMap(MapKeyToString(key.Interface()))
				if key_, err := self.unpack(path_, key, useStringMaps); err == nil {
					value_ := packedValue.MapIndex(key)
					if map_[MapKeyToString(key_)], err = self.unpack(path_, value_, useStringMaps); err != nil {
						return nil, fmt.Errorf("map value for %s", err.Error())
					}
				} else {
					return nil, fmt.Errorf("map key for %s", err.Error())
				}
			}
			return map_, nil
		} else {
			map_ := make(Map)
			keys := packedValue.MapKeys()
			for _, key := range keys {
				path_ := path.AppendMap(MapKeyToString(key.Interface()))
				if key_, err := self.unpack(path_, key, useStringMaps); err == nil {
					value_ := packedValue.MapIndex(key)
					if map_[key_], err = self.unpack(path_, value_, useStringMaps); err != nil {
						return nil, fmt.Errorf("map value for %s", err.Error())
					}
				} else {
					return nil, fmt.Errorf("map key for %s", err.Error())
				}
			}
			return map_, nil
		}

	case reflect.Struct:
		if useStringMaps {
			map_ := make(StringMap)
			for name, field := range self.newReflectFields(packedType) {
				value_ := packedValue.FieldByName(field.name)
				if value__, err := self.unpack(path.AppendField(field.name), value_, useStringMaps); err == nil {
					if !field.omitEmpty || !reflection.IsEmpty(value__) {
						map_[name] = value__
					}
				} else {
					return nil, err
				}
			}
			return map_, nil
		} else {
			map_ := make(Map)
			for name, field := range self.newReflectFields(packedType) {
				value_ := packedValue.FieldByName(field.name)
				if value__, err := self.unpack(path.AppendField(field.name), value_, useStringMaps); err == nil {
					if !field.omitEmpty || !reflection.IsEmpty(value__) {
						map_[name] = value__
					}
				} else {
					return nil, err
				}
			}
			return map_, nil
		}

	default:
		return nil, fmt.Errorf("%s is of unsupported type: %s", path.String(), packedType.String())
	}
}

//
// reflectField
//

type reflectField struct {
	name      string // actual field name
	omitEmpty bool
}

type reflectFields map[string]reflectField // key is user-defined name in tag

func (self *Reflector) newReflectFields(type_ reflect.Type) reflectFields {
	if reflectFields_, ok := self.reflectFieldsCache.Load(type_); ok {
		return reflectFields_.(reflectFields)
	}

	reflectFields_ := make(reflectFields)

	for _, structField := range reflection.GetStructFields(type_) {
		reflectField := reflectField{name: structField.Name}

		// Try tags in order
		tagged := false
		for _, structFieldTag := range self.StructFieldTags {
			if tag, ok := structField.Tag.Lookup(structFieldTag); ok {
				splitTag := strings.Split(tag, ",")
				if length := len(splitTag); length > 0 {
					name := splitTag[0]
					if name != "-" {
						if (length > 1) && (splitTag[1] == "omitempty") {
							reflectField.omitEmpty = true
						}
						reflectFields_[name] = reflectField
					}
				}

				tagged = true
				break
			}
		}

		if !tagged {
			if self.StructFieldNameMapper != nil {
				reflectFields_[self.StructFieldNameMapper(reflectField.name)] = reflectField
			} else {
				reflectFields_[reflectField.name] = reflectField
			}
		}
	}

	self.reflectFieldsCache.Store(type_, reflectFields_)

	return reflectFields_
}

func (self reflectFields) getField(structValue reflect.Value, name string) reflect.Value {
	return structValue.FieldByName(self[name].name)
}
