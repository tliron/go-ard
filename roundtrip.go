package ard

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

// Encodes and then decodes the value via a supported format.
//
// Supported formats are "yaml", "json", "xjson", "xml", "cbor", and "messagepack".
//
// While this function can be used to "canonicalize" values to ARD, it is
// generally be more efficient to call [ValidCopy] instead.
func Roundtrip(value Value, format string, reflector *Reflector) (Value, error) {
	switch format {
	case "yaml":
		return RoundtripYAML(value)

	case "json":
		return RoundtripJSON(value)

	case "xjson":
		return RoundtripXJSON(value, reflector)

	case "xml":
		return RoundtripXML(value, reflector)

	case "cbor":
		return RoundtripCBOR(value)

	case "messagepack":
		return RoundtripMessagePack(value)

	default:
		return nil, fmt.Errorf("unsupported format: %q", format)
	}
}

func RoundtripYAML(value Value) (Value, error) {
	var builder strings.Builder
	encoder := yaml.NewEncoder(&builder)
	if err := encoder.Encode(value); err == nil {
		value_, _, err := ReadYAML(strings.NewReader(builder.String()), false)
		return value_, err
	} else {
		return nil, err
	}
}

func RoundtripJSON(value Value) (Value, error) {
	var builder strings.Builder
	encoder := json.NewEncoder(&builder)
	if err := encoder.Encode(value); err == nil {
		return ReadJSON(strings.NewReader(builder.String()), true)
	} else {
		return nil, err
	}
}

func RoundtripXJSON(value Value, reflector *Reflector) (Value, error) {
	if value_, err := PrepareForEncodingXJSON(value, reflector); err == nil {
		var builder strings.Builder
		encoder := json.NewEncoder(&builder)
		if err := encoder.Encode(value_); err == nil {
			return ReadXJSON(strings.NewReader(builder.String()), true)
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func RoundtripXML(value Value, reflector *Reflector) (Value, error) {
	if value_, err := PrepareForEncodingXML(value, reflector); err == nil {
		var writer strings.Builder
		if _, err := writer.WriteString(xml.Header); err == nil {
			encoder := xml.NewEncoder(&writer)
			encoder.Indent("", "")
			if err := encoder.Encode(value_); err == nil {
				return ReadXML(strings.NewReader(writer.String()))
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func RoundtripCBOR(value Value) (Value, error) {
	if bytes, err := cbor.Marshal(value); err == nil {
		var value_ Value
		if err := cbor.Unmarshal(bytes, &value_); err == nil {
			return value_, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func RoundtripMessagePack(value Value) (Value, error) {
	var buffer bytes.Buffer
	encoder := NewMessagePackEncoder(&buffer)
	if err := encoder.Encode(value); err == nil {
		return ReadMessagePack(&buffer, false, true)
	} else {
		return nil, err
	}
}
