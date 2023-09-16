package ard

import (
	"bytes"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

// MessagePack decoder that supports "json" field tags.
func NewMessagePackDecoder(reader io.Reader) *msgpack.Decoder {
	decoder := msgpack.NewDecoder(reader)
	decoder.SetCustomStructTag("json")
	return decoder
}

// MessagePack encode that supports "json" field tags.
func NewMessagePackEncoder(writer io.Writer) *msgpack.Encoder {
	encoder := msgpack.NewEncoder(writer)
	encoder.SetCustomStructTag("json")
	return encoder
}

// Marshals MessagePack with support for "json" field tags.
func MarshalMessagePack(value any) ([]byte, error) {
	var bytes_ bytes.Buffer
	encoder := NewMessagePackEncoder(&bytes_)
	if err := encoder.Encode(value); err == nil {
		return bytes_.Bytes(), nil
	} else {
		return nil, err
	}
}

// Unmarshals MessagePack with support for "json" field tags.
func UnmarshalMessagePack(data []byte, value any) error {
	decoder := NewMessagePackDecoder(bytes.NewReader(data))
	return decoder.Decode(value)
}
