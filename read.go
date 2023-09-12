package ard

import (
	"bytes"
	contextpkg "context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/beevik/etree"
	"github.com/fxamacker/cbor/v2"
	"github.com/tliron/exturl"
	"github.com/tliron/kutil/util"
	"github.com/tliron/yamlkeys"
	"gopkg.in/yaml.v3"
)

// Reads and decodes supported formats to ARD.
//
// All resulting maps are guaranteed to be [Map] (and not [StringMap]).
//
// When locate is true will also attempt to return a [Locator] for the value if
// possible, otherwise will return nil.
//
// Supported formats are "yaml", "json", "xjson", "xml", "cbor", and "messagepack".
func Read(reader io.Reader, format string, locate bool) (Value, Locator, error) {
	switch format {
	case "yaml":
		return ReadYAML(reader, locate)

	case "json":
		value, err := ReadJSON(reader, false)
		return value, nil, err

	case "xjson":
		value, err := ReadXJSON(reader, false)
		return value, nil, err

	case "xml":
		value, err := ReadXML(reader)
		return value, nil, err

	case "cbor":
		value, err := ReadCBOR(reader, false)
		return value, nil, err

	case "messagepack":
		value, err := ReadMessagePack(reader, false, false)
		return value, nil, err

	default:
		return nil, nil, fmt.Errorf("unsupported format: %q", format)
	}
}

// Supported formats are "yaml", "json", "xjson", "xml", "cbor", and "messagepack".
func ReadURL(context contextpkg.Context, url exturl.URL, fallbackFormat string, locate bool) (Value, Locator, error) {
	if reader, err := url.Open(context); err == nil {
		reader = util.NewContextualReadCloser(context, reader)
		defer reader.Close()

		format := url.Format()
		if format == "" {
			format = fallbackFormat
		}

		return Read(reader, format, locate)
	} else {
		return nil, nil, err
	}
}

// Reads YAML from an [io.Reader] and decodes it into an ARD [Value].
// If more than one YAML document is present (i.e. separated by `---`)
// then only the first will be decoded with the remainder ignored.
//
// When locate is true will also return a [Locator] for the value,
// otherwise will return nil.
//
// Note that the YAML implementation uses the [yamlkeys] library, so
// that arbitrarily complex map keys are supported and decoded into
// ARD [Map]. If you need to manipulate these maps you should use the
// [yamlkeys] utility functions.
func ReadYAML(reader io.Reader, locate bool) (Value, Locator, error) {
	var node yaml.Node
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&node); err == nil {
		if value, err := yamlkeys.DecodeNode(&node); err == nil {
			var locator Locator
			if locate {
				locator = NewYAMLLocator(&node)
			}
			return value, locator, nil
		} else {
			return nil, nil, err
		}
	} else {
		return nil, nil, yamlkeys.WrapWithDecodeError(err)
	}
}

func ReadAllYAML(reader io.Reader) (List, error) {
	return yamlkeys.DecodeAll(reader)
}

// Reads JSON from an [io.Reader] and decodes it into an ARD [Value].
//
// If useStringMaps is true returns maps as [StringMap], otherwise they will
// be [Map].
func ReadJSON(reader io.Reader, useStringMaps bool) (Value, error) {
	var value Value
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&value); err == nil {
		// The JSON decoder uses StringMaps, not Maps
		if !useStringMaps {
			value, _ = ConvertStringMapsToMaps(value)
		}
		return value, nil
	} else {
		return nil, err
	}
}

// Reads JSON from an [io.Reader] and decodes it into an ARD [Value] while
// interpreting the xjson extensions.
//
// If useStringMaps is true returns maps as [StringMap], otherwise they will
// be [Map].
func ReadXJSON(reader io.Reader, useStringMaps bool) (Value, error) {
	var value Value
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&value); err == nil {
		value, _ = UnpackXJSON(value, useStringMaps)
		return value, nil
	} else {
		return nil, err
	}
}

// Reads XML from an [io.Reader] and decodes it into an ARD [Value].
func ReadXML(reader io.Reader) (Value, error) {
	document := etree.NewDocument()
	if _, err := document.ReadFrom(reader); err == nil {
		elements := document.ChildElements()
		if length := len(elements); length == 1 {
			value, err := UnpackXML(elements[0])
			return value, err
		} else {
			return nil, fmt.Errorf("unsupported XML: %d documents", length)
		}
	} else {
		return nil, err
	}
}

// Reads CBOR from an [io.Reader] and decodes it into an ARD [Value].
func ReadCBOR(reader io.Reader, base64 bool) (Value, error) {
	if base64 {
		var err error
		if reader, err = fromBase64(reader); err != nil {
			return nil, err
		}
	}

	var value Value
	decoder := cbor.NewDecoder(reader)
	if err := decoder.Decode(&value); err == nil {
		return value, nil
	} else {
		return nil, err
	}
}

// Reads CBOR from an [io.Reader] and decodes it into an ARD [Value].
//
// If useStringMaps is true returns maps as [StringMap], otherwise they will
// be [Map].
func ReadMessagePack(reader io.Reader, base64 bool, useStringMaps bool) (Value, error) {
	if base64 {
		var err error
		if reader, err = fromBase64(reader); err != nil {
			return nil, err
		}
	}

	var value Value
	decoder := NewMessagePackDecoder(reader)
	if err := decoder.Decode(&value); err == nil {
		// The MessagePack decoder uses StringMaps, not Maps
		if !useStringMaps {
			value, _ = ConvertStringMapsToMaps(value)
		}
		return value, nil
	} else {
		return nil, err
	}
}

func fromBase64(reader io.Reader) (io.Reader, error) {
	if b64, err := io.ReadAll(reader); err == nil {
		if code, err := util.FromBase64(util.BytesToString(b64)); err == nil {
			return bytes.NewReader(code), nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
