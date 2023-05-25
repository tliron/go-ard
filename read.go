package ard

import (
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

func Read(reader io.Reader, format string, locate bool) (Value, Locator, error) {
	switch format {
	case "yaml", "":
		return ReadYAML(reader, locate)

	case "json":
		return ReadJSON(reader, locate)

	case "cjson":
		return ReadCompatibleJSON(reader, locate)

	case "xml":
		return ReadCompatibleXML(reader, locate)

	case "cbor":
		return ReadCBOR(reader)

	case "messagepack":
		return ReadMessagePack(reader)

	default:
		return nil, nil, fmt.Errorf("unsupported format: %q", format)
	}
}

func ReadURL(context contextpkg.Context, url exturl.URL, locate bool) (Value, Locator, error) {
	if reader, err := url.Open(context); err == nil {
		reader = util.NewContextualReadCloser(context, reader)
		defer reader.Close()
		return Read(reader, url.Format(), locate)
	} else {
		return nil, nil, err
	}
}

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

func ReadJSON(reader io.Reader, locate bool) (Value, Locator, error) {
	var value Value
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&value); err == nil {
		// The JSON decoder uses StringMaps, not Maps
		value, _ := NormalizeMaps(value)
		return value, nil, nil
	} else {
		return nil, nil, err
	}
}

func ReadCompatibleJSON(reader io.Reader, locate bool) (Value, Locator, error) {
	var value Value
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&value); err == nil {
		value, _ = FromCompatibleJSON(value)
		return value, nil, nil
	} else {
		return nil, nil, err
	}
}

func ReadCompatibleXML(reader io.Reader, locate bool) (Value, Locator, error) {
	document := etree.NewDocument()
	if _, err := document.ReadFrom(reader); err == nil {
		elements := document.ChildElements()
		if length := len(elements); length == 1 {
			value, err := FromCompatibleXML(elements[0])
			return value, nil, err
		} else {
			return nil, nil, fmt.Errorf("unsupported XML: %d documents", length)
		}
	} else {
		return nil, nil, err
	}
}

func ReadCBOR(reader io.Reader) (Value, Locator, error) {
	var value Value
	decoder := cbor.NewDecoder(reader)
	if err := decoder.Decode(&value); err == nil {
		return value, nil, nil
	} else {
		return nil, nil, err
	}
}

func ReadMessagePack(reader io.Reader) (Value, Locator, error) {
	var value Value
	decoder := util.NewMessagePackDecoder(reader)
	if err := decoder.Decode(&value); err == nil {
		// The MessagePack decoder uses StringMaps, not Maps
		value, _ := NormalizeMaps(value)
		return value, nil, nil
	} else {
		return nil, nil, err
	}
}
