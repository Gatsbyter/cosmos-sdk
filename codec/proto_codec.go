package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
)

// ProtoCodec defines a codec that utilizes Protobuf for both binary and JSON
// encoding.
type ProtoCodec struct{}

func NewProtoCodec() Marshaler {
	return &ProtoCodec{}
}

func (pc *ProtoCodec) MarshalBinaryBare(o ProtoMarshaler) ([]byte, error) {
	return o.Marshal()
}

func (pc *ProtoCodec) MustMarshalBinaryBare(o ProtoMarshaler) []byte {
	bz, err := pc.MarshalBinaryBare(o)
	if err != nil {
		panic(err)
	}

	return bz
}

func (pc *ProtoCodec) MarshalBinaryLengthPrefixed(o ProtoMarshaler) ([]byte, error) {
	bz, err := pc.MarshalBinaryBare(o)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := encodeUvarint(buf, uint64(o.Size())); err != nil {
		return nil, err
	}

	if _, err := buf.Write(bz); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (pc *ProtoCodec) MustMarshalBinaryLengthPrefixed(o ProtoMarshaler) []byte {
	bz, err := pc.MarshalBinaryLengthPrefixed(o)
	if err != nil {
		panic(err)
	}

	return bz
}

func (pc *ProtoCodec) UnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) error {
	return ptr.Unmarshal(bz)
}

func (pc *ProtoCodec) MustUnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) {
	if err := pc.UnmarshalBinaryBare(bz, ptr); err != nil {
		panic(err)
	}
}

func (pc *ProtoCodec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) error {
	size, n := binary.Uvarint(bz)
	if n < 0 {
		return fmt.Errorf("invalid number of bytes read from length-prefixed encoding: %d", n)
	}

	if size > uint64(len(bz)-n) {
		return fmt.Errorf("not enough bytes to read; want: %v, got: %v", size, len(bz)-n)
	} else if size < uint64(len(bz)-n) {
		return fmt.Errorf("too many bytes to read; want: %v, got: %v", size, len(bz)-n)
	}

	bz = bz[n:]
	return ptr.Unmarshal(bz)
}

func (pc *ProtoCodec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) {
	if err := pc.UnmarshalBinaryLengthPrefixed(bz, ptr); err != nil {
		panic(err)
	}
}

func (pc *ProtoCodec) MarshalJSON(o interface{}) ([]byte, error) { // nolint: stdmethods
	m, ok := o.(ProtoMarshaler)
	if !ok {
		return nil, fmt.Errorf("cannot protobuf JSON encode unsupported type: %T", o)
	}

	buf := new(bytes.Buffer)

	marshaler := &jsonpb.Marshaler{}
	if err := marshaler.Marshal(buf, m); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (pc *ProtoCodec) MustMarshalJSON(o interface{}) []byte {
	bz, err := pc.MarshalJSON(o)
	if err != nil {
		panic(err)
	}

	return bz
}

func (pc *ProtoCodec) UnmarshalJSON(bz []byte, ptr interface{}) error { // nolint: stdmethods
	m, ok := ptr.(ProtoMarshaler)
	if !ok {
		return fmt.Errorf("cannot protobuf JSON decode unsupported type: %T", ptr)
	}

	return jsonpb.Unmarshal(strings.NewReader(string(bz)), m)
}

func (pc *ProtoCodec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	if err := pc.UnmarshalJSON(bz, ptr); err != nil {
		panic(err)
	}
}
