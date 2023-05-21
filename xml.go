package ard

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"strconv"

	"github.com/beevik/etree"
	"github.com/tliron/kutil/util"
	"github.com/tliron/yamlkeys"
)

const (
	CompatibleXMLNilTag           = "nil"
	CompatibleXMLBytesTag         = "bytes"
	CompatibleXMLListTag          = "list"
	CompatibleXMLMapTag           = "map"
	CompatibleXMLMapEntryTag      = "entry"
	CompatibleXMLMapEntryKeyTag   = "key"
	CompatibleXMLMapEntryValueTag = "value"
)

func EnsureCompatibleXML(value Value) (Value, error) {
	if value_, err := Canonicalize(value); err == nil {
		return ToCompatibleXML(value_), nil
	} else {
		return nil, err
	}
}

func ToCompatibleXML(value any) any {
	if value == nil {
		return CompatibleXMLNil{}
	}

	if bytes, ok := value.([]byte); ok {
		return CompatibleXMLBytes{bytes}
	}

	value_ := reflect.ValueOf(value)

	switch value_.Type().Kind() {
	case reflect.Slice:
		length := value_.Len()
		slice := make([]any, length)
		for index := 0; index < length; index++ {
			v := value_.Index(index).Interface()
			slice[index] = ToCompatibleXML(v)
		}
		return CompatibleXMLList{slice}

	case reflect.Map:
		// Convert to slice of XMLMapEntry
		slice := make([]CompatibleXMLMapEntry, value_.Len())
		for index, key := range value_.MapKeys() {
			k := yamlkeys.KeyData(key.Interface())
			v := value_.MapIndex(key).Interface()
			slice[index] = CompatibleXMLMapEntry{
				key:   ToCompatibleXML(k),
				value: ToCompatibleXML(v),
			}
		}
		return CompatibleXMLMap{slice}
	}

	return value
}

func FromCompatibleXML(element *etree.Element) (any, error) {
	switch element.Tag {
	case CompatibleXMLNilTag:
		return nil, nil

	case CompatibleXMLListTag:
		children := element.ChildElements()
		list := make(List, len(children))
		for index, entry := range children {
			if entry_, err := FromCompatibleXML(entry); err == nil {
				list[index] = entry_
			} else {
				return nil, err
			}
		}
		return list, nil

	case CompatibleXMLMapTag:
		map_ := make(Map)
		for _, entry := range element.ChildElements() {
			if entry_, err := NewCompatibleXMLMapEntry(entry); err == nil {
				//fmt.Printf("%T\n", entry_.Key)
				map_[entry_.key] = entry_.value
			} else {
				return nil, err
			}
		}
		return map_, nil

	case "string":
		return element.Text(), nil

	case "int":
		if int_, err := strconv.ParseInt(element.Text(), 10, 64); err == nil {
			return int64(int_), nil
		} else {
			return nil, err
		}

	case "int64":
		return strconv.ParseInt(element.Text(), 10, 64)

	case "int32":
		return strconv.ParseInt(element.Text(), 10, 32)

	case "int8":
		return strconv.ParseInt(element.Text(), 10, 8)

	case "uint":
		if uint_, err := strconv.ParseUint(element.Text(), 10, 64); err == nil {
			return uint64(uint_), nil
		} else {
			return nil, err
		}

	case "uint64":
		return strconv.ParseUint(element.Text(), 10, 64)

	case "uint32":
		return strconv.ParseUint(element.Text(), 10, 32)

	case "uint8":
		return strconv.ParseUint(element.Text(), 10, 8)

	case "float64":
		return strconv.ParseFloat(element.Text(), 64)

	case "float32":
		return strconv.ParseFloat(element.Text(), 32)

	case "bool":
		return strconv.ParseBool(element.Text())

	case "bytes":
		return util.FromBase64(element.Text())

	default:
		return nil, fmt.Errorf("element has unsupported tag: %s", xmlElementToString(element))
	}
}

//
// CompatibleXMLList
//

type CompatibleXMLList struct {
	Entries []any
}

// xml.Marshaler interface
func (self CompatibleXMLList) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	listStart := xml.StartElement{Name: xml.Name{Local: CompatibleXMLListTag}}

	if err := encoder.EncodeToken(listStart); err == nil {
		if err := encoder.Encode(self.Entries); err == nil {
			return encoder.EncodeToken(listStart.End())
		} else {
			return err
		}
	} else {
		return err
	}
}

//
// CompatibleXMLNil
//

type CompatibleXMLNil struct{}

var nilStart = xml.StartElement{Name: xml.Name{Local: CompatibleXMLNilTag}}

// xml.Marshaler interface
func (self CompatibleXMLNil) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if err := encoder.EncodeToken(nilStart); err == nil {
		return encoder.EncodeToken(nilStart.End())
	} else {
		return err
	}
}

//
// CompatibleXMLBytes
//

type CompatibleXMLBytes struct {
	bytes []byte
}

var bytesStart = xml.StartElement{Name: xml.Name{Local: CompatibleXMLBytesTag}}

// xml.Marshaler interface
func (self CompatibleXMLBytes) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if err := encoder.EncodeToken(bytesStart); err == nil {
		if err := encoder.EncodeToken(xml.CharData(util.StringToBytes(util.ToBase64(self.bytes)))); err == nil {
			return encoder.EncodeToken(bytesStart.End())
		} else {
			return err
		}
	} else {
		return err
	}
}

//
// CompatibleXMLMap
//

type CompatibleXMLMap struct {
	entries []CompatibleXMLMapEntry
}

var mapStart = xml.StartElement{Name: xml.Name{Local: CompatibleXMLMapTag}}

// xml.Marshaler interface
func (self CompatibleXMLMap) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if err := encoder.EncodeToken(mapStart); err == nil {
		if err := encoder.Encode(self.entries); err == nil {
			return encoder.EncodeToken(mapStart.End())
		} else {
			return err
		}
	} else {
		return err
	}
}

//
// CompatibleXMLMapEntry
//

type CompatibleXMLMapEntry struct {
	key   any
	value any
}

var mapEntryStart = xml.StartElement{Name: xml.Name{Local: CompatibleXMLMapEntryTag}}
var keyStart = xml.StartElement{Name: xml.Name{Local: CompatibleXMLMapEntryKeyTag}}
var valueStart = xml.StartElement{Name: xml.Name{Local: CompatibleXMLMapEntryValueTag}}

// xml.Marshaler interface
func (self CompatibleXMLMapEntry) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if err := encoder.EncodeToken(mapEntryStart); err == nil {
		if err := encoder.EncodeToken(keyStart); err == nil {
			if err := encoder.Encode(self.key); err == nil {
				if err := encoder.EncodeToken(keyStart.End()); err == nil {
					if err := encoder.EncodeToken(valueStart); err == nil {
						if err := encoder.Encode(self.value); err == nil {
							if err := encoder.EncodeToken(valueStart.End()); err == nil {
								return encoder.EncodeToken(mapEntryStart.End())
							} else {
								return err
							}
						} else {
							return err
						}
					} else {
						return err
					}
				} else {
					return err
				}
			} else {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}
}

func NewCompatibleXMLMapEntry(element *etree.Element) (CompatibleXMLMapEntry, error) {
	var self CompatibleXMLMapEntry

	if element.Tag == CompatibleXMLMapEntryTag {
		for _, child := range element.ChildElements() {
			switch child.Tag {
			case CompatibleXMLMapEntryKeyTag:
				if key, err := getXmlElementSingleChild(child); err == nil {
					self.key = key
				} else {
					return self, err
				}

			case CompatibleXMLMapEntryValueTag:
				if value, err := getXmlElementSingleChild(child); err == nil {
					self.value = value
				} else {
					return self, err
				}

			default:
				return self, fmt.Errorf("element has unsupported tag: %s", xmlElementToString(element))
			}
		}
	}

	return self, nil
}

// Utilities

func xmlElementToString(element *etree.Element) string {
	document := etree.NewDocument()
	document.SetRoot(element)
	if s, err := document.WriteToString(); err == nil {
		return s
	} else {
		return element.GetPath()
	}
}

func getXmlElementSingleChild(element *etree.Element) (any, error) {
	children := element.ChildElements()
	length := len(children)
	if length == 1 {
		return FromCompatibleXML(children[0])
	} else if length == 0 {
		return nil, nil
	} else {
		return nil, fmt.Errorf("element has more than one child: %s", xmlElementToString(element))
	}
}
