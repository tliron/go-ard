package ard

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tliron/kutil/util"
)

func Decode(code string, format string, locate bool) (Value, Locator, error) {
	switch format {
	case "yaml", "":
		return DecodeYAML(code, locate)

	case "json":
		return DecodeJSON(code, locate)

	case "cjson":
		return DecodeCompatibleJSON(code, locate)

	case "xml":
		return DecodeCompatibleXML(code, locate)

	case "cbor":
		return DecodeCBOR(code)

	case "messagepack":
		return DecodeMessagePack(code)

	default:
		return nil, nil, fmt.Errorf("unsupported format: %q", format)
	}
}

func DecodeYAML(code string, locate bool) (Value, Locator, error) {
	return ReadYAML(strings.NewReader(code), locate)
}

func DecodeJSON(code string, locate bool) (Value, Locator, error) {
	return ReadJSON(strings.NewReader(code), locate)
}

func DecodeCompatibleJSON(code string, locate bool) (Value, Locator, error) {
	return ReadCompatibleJSON(strings.NewReader(code), locate)
}

func DecodeCompatibleXML(code string, locate bool) (Value, Locator, error) {
	return ReadCompatibleXML(strings.NewReader(code), locate)
}

// The code should be in Base64
func DecodeCBOR(code string) (Value, Locator, error) {
	if bytes_, err := util.FromBase64(code); err == nil {
		return ReadCBOR(bytes.NewReader(bytes_))
	} else {
		return nil, nil, err
	}
}

// The code should be in Base64
func DecodeMessagePack(code string) (Value, Locator, error) {
	if bytes_, err := util.FromBase64(code); err == nil {
		return ReadMessagePack(bytes.NewReader(bytes_))
	} else {
		return nil, nil, err
	}
}
