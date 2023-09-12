package ard

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/fxamacker/cbor/v2"
	"github.com/tliron/kutil/util"
	"github.com/vmihailenco/msgpack/v5"
	"gopkg.in/yaml.v3"
)

// Validates that data is of a supported format. Returns nil if valid, otherwise
// it will be the validation error.
//
// While you can use [Read] to validate data, this function is optimized to be
// more lightweight. On other hand, this function does not do any schema validation
// (for example for XML), so if this function returns no error it does not
// guarantee that [Read] would also not return an error.
func Validate(code []byte, format string) error {
	switch format {
	case "yaml":
		return ValidateYAML(code)

	case "json", "xjson":
		return ValidateJSON(code)

	case "xml":
		return ValidateXML(code)

	case "cbor":
		return ValidateCBOR(code, false)

	case "messagepack":
		return ValidateMessagePack(code, false)

	default:
		return fmt.Errorf("unsupported format: %q", format)
	}
}

func ValidateYAML(code []byte) error {
	decoder := yaml.NewDecoder(bytes.NewReader(code))
	// Note: decoder.Decode will only decode the first document it finds
	for {
		var node yaml.Node
		if err := decoder.Decode(&node); err != nil {
			if err == io.EOF {
				return nil
			} else {
				return err
			}
		}
	}
}

func ValidateJSON(code []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(code))
	// Note: decoder.Decode will only decode the first element it finds
	for {
		if _, err := decoder.Token(); err != nil {
			if err == io.EOF {
				return nil
			} else {
				return err
			}
		}
	}
}

func ValidateXML(code []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(code))
	// Note: decoder.Decode will only decode the first element it finds
	for {
		if _, err := decoder.Token(); err != nil {
			if err == io.EOF {
				return nil
			} else {
				return err
			}
		}
	}
}

func ValidateCBOR(code []byte, base64 bool) error {
	if base64 {
		var err error
		if code, err = util.FromBase64(util.BytesToString(code)); err != nil {
			return err
		}
	}

	var value any
	if err := cbor.Unmarshal(code, &value); err == nil {
		return nil
	} else {
		return err
	}
}

func ValidateMessagePack(code []byte, base64 bool) error {
	if base64 {
		var err error
		if code, err = util.FromBase64(util.BytesToString(code)); err != nil {
			return err
		}
	}

	var value any
	if err := msgpack.Unmarshal(code, &value); err == nil {
		return nil
	} else {
		return err
	}
}
