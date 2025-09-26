package ard

import (
	"bytes"
	"io"

	"github.com/tliron/go-kutil/util"
)

type conversionMode int

const (
	noConversion            conversionMode = 0
	convertStringMapsToMaps conversionMode = 1
	convertMapsToStringMaps conversionMode = 2
)

func readBase64(reader io.Reader) (io.Reader, error) {
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
