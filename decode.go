package ard

import (
	"bytes"
	"fmt"
)

// Decodes supported formats to ARD.
//
// All resulting maps are guaranteed to be [Map] (and not [StringMap]).
//
// Supported formats are "yaml", "json", "xjson", "xml", "cbor", and "messagepack".
func Decode(code []byte, format string, locate bool) (Value, Locator, error) {
	switch format {
	case "yaml":
		return DecodeYAML(code, locate)

	case "json":
		value, err := DecodeJSON(code, false)
		return value, nil, err

	case "xjson":
		value, err := DecodeXJSON(code, false)
		return value, nil, err

	case "xml":
		value, err := DecodeXML(code)
		return value, nil, err

	case "cbor":
		value, err := DecodeCBOR(code, false)
		return value, nil, err

	case "messagepack":
		value, err := DecodeMessagePack(code, false, false)
		return value, nil, err

	default:
		return nil, nil, fmt.Errorf("unsupported format: %q", format)
	}
}

func DecodeYAML(code []byte, locate bool) (Value, Locator, error) {
	return ReadYAML(bytes.NewReader(code), locate)
}

func DecodeJSON(code []byte, useStringMaps bool) (Value, error) {
	return ReadJSON(bytes.NewReader(code), useStringMaps)
}

func DecodeXJSON(code []byte, useStringMaps bool) (Value, error) {
	return ReadXJSON(bytes.NewReader(code), useStringMaps)
}

func DecodeXML(code []byte) (Value, error) {
	return ReadXML(bytes.NewReader(code))
}

func DecodeCBOR(code []byte, base64 bool) (Value, error) {
	return ReadCBOR(bytes.NewReader(code), base64)
}

func DecodeMessagePack(code []byte, base64 bool, useStringMaps bool) (Value, error) {
	return ReadMessagePack(bytes.NewReader(code), base64, useStringMaps)
}
