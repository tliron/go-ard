package ard

import (
	"bytes"
	"fmt"
	templatepkg "text/template"
)

// Decodes supported formats to an ARD [Value].
//
// All resulting maps are guaranteed to be [Map] (and not [StringMap]).
//
// If locate is true then a [Locator] will be returned if possible.
// Currently only YAML decoding supports this feature.
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

// Convenience function to parse and render a template and then decode it.
//
// Would only be useful for text-based formats, so not CBOR and MessagePack.
//
// See text/template.
func DecodeTemplate(template string, data any, format string, locate bool) (Value, Locator, error) {
	if template_, err := templatepkg.New("ard").Parse(template); err == nil {
		var buffer bytes.Buffer
		if err := template_.Execute(&buffer, data); err == nil {
			return Read(&buffer, format, false)
		} else {
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}
}

// Decodes YAML to an ARD [Value]. If more than one YAML document is
// present (i.e. separated by `---`) then only the first will be decoded
// with the remainder ignored.
//
// When locate is true will also return a [Locator] for the value,
// otherwise will return nil.
//
// Note that the YAML implementation uses the [yamlkeys] library, so
// that arbitrarily complex map keys are supported and decoded into
// ARD [Map]. If you need to manipulate these maps you should use the
// [yamlkeys] utility functions.
func DecodeYAML(code []byte, locate bool) (Value, Locator, error) {
	return ReadYAML(bytes.NewReader(code), locate)
}

// Decodes JSON to an ARD [Value].
//
// If useStringMaps is true returns maps as [StringMap], otherwise they will
// be [Map].
func DecodeJSON(code []byte, useStringMaps bool) (Value, error) {
	return ReadJSON(bytes.NewReader(code), useStringMaps)
}

// Decodes JSON to an ARD [Value] while interpreting the XJSON extensions.
//
// If useStringMaps is true returns maps as [StringMap], otherwise they will
// be [Map].
func DecodeXJSON(code []byte, useStringMaps bool) (Value, error) {
	return ReadXJSON(bytes.NewReader(code), useStringMaps)
}

// Decodes XML to an ARD [Value].
//
// A specific schema is expected (currently undocumented).
func DecodeXML(code []byte) (Value, error) {
	return ReadXML(bytes.NewReader(code))
}

// Decodes CBOR to an ARD [Value].
//
// If base64 is true then the data will first be fully read and decoded from
// Base64 to bytes.
func DecodeCBOR(code []byte, base64 bool) (Value, error) {
	return ReadCBOR(bytes.NewReader(code), base64)
}

// Reads MessagePack from an [io.Reader] and decodes it to an ARD [Value].
//
// If base64 is true then the data will first be fully read and decoded from
// Base64 to bytes.
//
// If useStringMaps is true returns maps as [StringMap], otherwise they will
// be [Map].
func DecodeMessagePack(code []byte, base64 bool, useStringMaps bool) (Value, error) {
	return ReadMessagePack(bytes.NewReader(code), base64, useStringMaps)
}
