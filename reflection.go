package ard

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/tliron/kutil/reflection"
)

type StructFieldNameMapperFunc func(fieldName string) string

type FromARD interface {
	FromARD(reflector *Reflector) (any, error)
}

type ToARD interface {
	ToARD(reflector *Reflector) (any, error)
}

//
// Reflector
//

var defaultStructFieldTags = []string{"ard", "yaml", "json"}

type Reflector struct {
	IgnoreMissingStructFields bool
	NilMeansZero              bool
	StructFieldTags           []string // in order
	StructFieldNameMapper     StructFieldNameMapperFunc

	reflectFieldsCache sync.Map
}

func NewReflector() *Reflector {
	return &Reflector{StructFieldTags: defaultStructFieldTags}
}

// Fills in Go struct fields from ARD maps
func (self *Reflector) Pack(value Value, packedValuePtr any) error {
	packedValuePtr_ := reflect.ValueOf(packedValuePtr)
	if packedValuePtr_.Kind() == reflect.Pointer {
		return self.PackReflect(value, packedValuePtr_)
	} else {
		return fmt.Errorf("not a pointer: %T", packedValuePtr)
	}
}

func (self *Reflector) PackReflect(value Value, packedValue reflect.Value) error {
	packedType := packedValue.Type()

	// Dereference pointers
	for packedType.Kind() == reflect.Pointer {
		if value == nil {
			return nil
		}

		packedType = packedType.Elem()
		if packedValue.IsNil() {
			// Allocate zero value on heap
			packedValue.Set(reflect.New(packedType))
		}
		packedValue = packedValue.Elem()
	}

	switch value_ := value.(type) {
	case nil:
		if self.NilMeansZero {
			packedValue.Set(reflect.Zero(packedType))
		} else {
			kind := packedValue.Kind()
			if (kind != reflect.Map) && (kind != reflect.Slice) {
				return fmt.Errorf("not a pointer, map, or slice: %s", packedType.String())
			}
		}

	case string:
		if packedValue.Kind() == reflect.String {
			packedValue.SetString(value_)
		} else {
			return fmt.Errorf("not a string: %s", packedType.String())
		}

	case bool:
		if packedValue.Kind() == reflect.Bool {
			packedValue.SetBool(value_)
		} else {
			return fmt.Errorf("not a bool: %s", packedType.String())
		}

	case int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint:
		if reflection.IsInteger(packedValue.Kind()) {
			packedValue.SetInt(ToInt64(value_))
		} else if reflection.IsUInteger(packedValue.Kind()) {
			packedValue.SetUint(ToUInt64(value_))
		} else {
			return fmt.Errorf("not an integer: %s", packedType.String())
		}

	case float64, float32:
		if reflection.IsFloat(packedValue.Kind()) {
			packedValue.SetFloat(ToFloat64(value_))
		} else {
			return fmt.Errorf("not a float: %s", packedType.String())
		}

	case []byte, time.Time: // as-is values
		if packedType == reflect.TypeOf(value_) {
			packedValue.Set(reflect.ValueOf(value_))
		} else {
			return fmt.Errorf("not a %T: %s", value, packedType.String())
		}

	case List:
		if packedValue.Kind() == reflect.Slice {
			elemType := packedType.Elem()
			length := len(value_)
			list := reflect.MakeSlice(reflect.SliceOf(elemType), length, length)
			for index, elem := range value_ {
				if err := self.PackReflect(elem, list.Index(index)); err != nil {
					return fmt.Errorf("slice element %d %s", index, err.Error())
				}
			}
			packedValue.Set(list)
		} else {
			return fmt.Errorf("not a slice: %s", packedType.String())
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
				if err := self.PackReflect(k, k_); err == nil {
					v_ := reflect.New(valueType)
					if err := self.PackReflect(v, v_); err == nil {
						packedValue.SetMapIndex(k_.Elem(), v_.Elem())
					} else {
						return fmt.Errorf("map value %s", err)
					}
				} else {
					return fmt.Errorf("map key %s", err)
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

			reflectFields := self.NewReflectFields(packedType)
			for k, v := range value_ {
				if k_, ok := k.(string); ok {
					if err := self.setStructField(packedValue, k_, v, reflectFields); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("key %q not a string: %T", k, k)
				}
			}

		default:
			return fmt.Errorf("not a map or struct: %s", packedType.String())
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
				if err := self.PackReflect(k, k_); err == nil {
					v_ := reflect.New(valueType)
					if err := self.PackReflect(v, v_); err == nil {
						packedValue.SetMapIndex(k_.Elem(), v_.Elem())
					} else {
						return fmt.Errorf("map value %s", err)
					}
				} else {
					return fmt.Errorf("map key %s", err)
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

			reflectFields := self.NewReflectFields(packedType)
			for k, v := range value_ {
				if err := self.setStructField(packedValue, k, v, reflectFields); err != nil {
					return err
				}
			}

		default:
			return fmt.Errorf("not a map or struct: %s", packedType.String())
		}

	default:
		return fmt.Errorf("unsupported type: %s", packedType.String())
	}

	return nil
}

// Converts Go structs to ARD maps
func (self *Reflector) Unpack(packedValue any, useStringMaps bool) (Value, error) {
	return self.UnpackReflect(reflect.ValueOf(packedValue), useStringMaps)
}

var time_ time.Time
var timeType = reflect.TypeOf(time_)

func (self *Reflector) UnpackReflect(packedValue reflect.Value, useStringMaps bool) (Value, error) {
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
			if list[index], err = self.UnpackReflect(value_, useStringMaps); err != nil {
				return nil, fmt.Errorf("list element %d %s", index, err.Error())
			}
		}
		return list, nil

	case reflect.Map:
		if useStringMaps {
			map_ := make(StringMap)
			keys := packedValue.MapKeys()
			for _, key := range keys {
				if key_, err := self.UnpackReflect(key, useStringMaps); err == nil {
					value_ := packedValue.MapIndex(key)
					if map_[MapKeyToString(key_)], err = self.UnpackReflect(value_, useStringMaps); err != nil {
						return nil, fmt.Errorf("map value %q %s", key_, err.Error())
					}
				} else {
					return nil, fmt.Errorf("map key %q %s", key, err.Error())
				}
			}
			return map_, nil
		} else {
			map_ := make(Map)
			keys := packedValue.MapKeys()
			for _, key := range keys {
				if key_, err := self.UnpackReflect(key, useStringMaps); err == nil {
					value_ := packedValue.MapIndex(key)
					if map_[key_], err = self.UnpackReflect(value_, useStringMaps); err != nil {
						return nil, fmt.Errorf("map value %q %s", key_, err.Error())
					}
				} else {
					return nil, fmt.Errorf("map key %q %s", key, err.Error())
				}
			}
			return map_, nil
		}

	case reflect.Struct:
		if useStringMaps {
			map_ := make(StringMap)
			for name, field := range self.NewReflectFields(packedType) {
				value_ := packedValue.FieldByName(field.name)
				if value__, err := self.UnpackReflect(value_, useStringMaps); err == nil {
					if !field.omitEmpty || !reflection.IsEmpty(value__) {
						map_[name] = value__
					}
				} else {
					return nil, fmt.Errorf("struct field %q %s", field.name, err.Error())
				}
			}
			return map_, nil
		} else {
			map_ := make(Map)
			for name, field := range self.NewReflectFields(packedType) {
				value_ := packedValue.FieldByName(field.name)
				if value__, err := self.UnpackReflect(value_, useStringMaps); err == nil {
					if !field.omitEmpty || !reflection.IsEmpty(value__) {
						map_[name] = value__
					}
				} else {
					return nil, fmt.Errorf("struct field %q %s", field.name, err.Error())
				}
			}
			return map_, nil
		}

	default:
		return nil, fmt.Errorf("unsupported type: %s", packedType.String())
	}
}

func (self *Reflector) setStructField(structValue reflect.Value, fieldName string, value Value, fieldNames ReflectFields) error {
	field := fieldNames.GetField(structValue, fieldName)

	if !field.IsValid() {
		if self.IgnoreMissingStructFields {
			return nil
		} else {
			return fmt.Errorf("no %q field", fieldName)
		}
	}

	if !field.CanSet() {
		return fmt.Errorf("field %q cannot be set", fieldName)
	}

	if err := self.PackReflect(value, field); err == nil {
		return nil
	} else {
		return fmt.Errorf("field %q %s", fieldName, err.Error())
	}
}

//
// ReflectField
//

type ReflectField struct {
	name      string
	omitEmpty bool
}

type ReflectFields map[string]ReflectField // ARD name

func (self *Reflector) NewReflectFields(type_ reflect.Type) ReflectFields {
	if reflectFields, ok := self.reflectFieldsCache.Load(type_); ok {
		return reflectFields.(ReflectFields)
	}

	reflectFields := make(ReflectFields)

	for _, structField := range reflection.GetStructFields(type_) {
		reflectField := ReflectField{name: structField.Name}

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
						reflectFields[name] = reflectField
					}
				}

				tagged = true
				break
			}
		}

		if !tagged {
			if self.StructFieldNameMapper != nil {
				reflectFields[self.StructFieldNameMapper(reflectField.name)] = reflectField
			} else {
				reflectFields[reflectField.name] = reflectField
			}
		}
	}

	self.reflectFieldsCache.Store(type_, reflectFields)

	return reflectFields
}

func (self ReflectFields) GetField(structValue reflect.Value, name string) reflect.Value {
	return structValue.FieldByName(self[name].name)
}
