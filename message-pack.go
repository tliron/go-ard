package ard

import (
	"bytes"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

func NewMessagePackDecoder(reader io.Reader) *msgpack.Decoder {
	decoder := msgpack.NewDecoder(reader)
	decoder.SetCustomStructTag("json")
	return decoder
}

func NewMessagePackEncoder(writer io.Writer) *msgpack.Encoder {
	encoder := msgpack.NewEncoder(writer)
	encoder.SetCustomStructTag("json")
	return encoder
}

func MarshalMessagePack(value any) ([]byte, error) {
	var bytes_ bytes.Buffer
	encoder := NewMessagePackEncoder(&bytes_)
	if err := encoder.Encode(value); err == nil {
		return bytes_.Bytes(), nil
	} else {
		return nil, err
	}
}

func UnmarshalMessagePack(data []byte, value any) error {
	decoder := NewMessagePackDecoder(bytes.NewReader(data))
	return decoder.Decode(value)
}
