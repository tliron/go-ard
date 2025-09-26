package ard

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"strconv"

	"github.com/beevik/etree"
	"github.com/tliron/go-kutil/util"
	"github.com/tliron/yamlkeys"
)

const (
	XMLNilTag           = "nil"
	XMLBytesTag         = "bytes"
	XMLListTag          = "list"
	XMLMapTag           = "map"
	XMLMapEntryTag      = "entry"
	XMLMapEntryKeyTag   = "key"
	XMLMapEntryValueTag = "value"
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
func PrepareForEncodingXML(value Value, inPlace bool, reflector *Reflector) (any, error) {
	if !inPlace {
		var err error
		if value, err = ValidCopy(value, reflector); err != nil {
			return nil, err
		}
	}

	return PackXML(value), nil
}

func PackXML(value Value) any {
	if value == nil {
		return XMLNil{}
	}

	if bytes, ok := value.([]byte); ok {
		return XMLBytes{bytes}
	}

	value_ := reflect.ValueOf(value)

	switch value_.Type().Kind() {
	case reflect.Slice:
		length := value_.Len()
		slice := make([]any, length)
		for index := 0; index < length; index++ {
			v := value_.Index(index).Interface()
			slice[index] = PackXML(v)
		}
		return XMLList{slice}

	case reflect.Map:
		// Convert to slice of XMLMapEntry
		slice := make([]XMLMapEntry, value_.Len())
		for index, key := range value_.MapKeys() {
			k := yamlkeys.KeyData(key.Interface())
			v := value_.MapIndex(key).Interface()
			slice[index] = XMLMapEntry{
				key:   PackXML(k),
				value: PackXML(v),
			}
		}
		return XMLMap{slice}
	}

	return value
}

func UnpackXML(element *etree.Element) (Value, error) {
	switch element.Tag {
	case XMLNilTag:
		return nil, nil

	case XMLListTag:
		children := element.ChildElements()
		list := make(List, len(children))
		for index, entry := range children {
			if entry_, err := UnpackXML(entry); err == nil {
				list[index] = entry_
			} else {
				return nil, err
			}
		}
		return list, nil

	case XMLMapTag:
		map_ := make(Map)
		for _, entry := range element.ChildElements() {
			if entry_, err := NewXMLMapEntry(entry); err == nil {
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
// XMLList
//

type XMLList struct {
	Entries []any
}

// ([xml.Marshaler] interface)
func (self XMLList) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	listStart := xml.StartElement{Name: xml.Name{Local: XMLListTag}}

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
// XMLNil
//

type XMLNil struct{}

var nilStart = xml.StartElement{Name: xml.Name{Local: XMLNilTag}}

// ([xml.Marshaler] interface)
func (self XMLNil) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if err := encoder.EncodeToken(nilStart); err == nil {
		return encoder.EncodeToken(nilStart.End())
	} else {
		return err
	}
}

//
// XMLBytes
//

type XMLBytes struct {
	bytes []byte
}

var bytesStart = xml.StartElement{Name: xml.Name{Local: XMLBytesTag}}

// ([xml.Marshaler] interface)
func (self XMLBytes) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
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
// XMLMap
//

type XMLMap struct {
	entries []XMLMapEntry
}

var mapStart = xml.StartElement{Name: xml.Name{Local: XMLMapTag}}

// ([xml.Marshaler] interface)
func (self XMLMap) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
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
// XMLMapEntry
//

type XMLMapEntry struct {
	key   any
	value any
}

var mapEntryStart = xml.StartElement{Name: xml.Name{Local: XMLMapEntryTag}}
var keyStart = xml.StartElement{Name: xml.Name{Local: XMLMapEntryKeyTag}}
var valueStart = xml.StartElement{Name: xml.Name{Local: XMLMapEntryValueTag}}

// ([xml.Marshaler] interface)
func (self XMLMapEntry) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
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

func NewXMLMapEntry(element *etree.Element) (XMLMapEntry, error) {
	var self XMLMapEntry

	if element.Tag == XMLMapEntryTag {
		for _, child := range element.ChildElements() {
			switch child.Tag {
			case XMLMapEntryKeyTag:
				if key, err := getXmlElementSingleChild(child); err == nil {
					self.key = key
				} else {
					return self, err
				}

			case XMLMapEntryValueTag:
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
		return UnpackXML(children[0])
	} else if length == 0 {
		return nil, nil
	} else {
		return nil, fmt.Errorf("element has more than one child: %s", xmlElementToString(element))
	}
}
