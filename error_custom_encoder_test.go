package msgpack_test

import (
	"bytes"
	"testing"

	msgpack "github.com/Basekick-Labs/msgpack/v6"
)

type pointerEncoderError string

func (e pointerEncoderError) Error() string { return string(e) }

func (e *pointerEncoderError) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeString("custom:" + string(*e))
}

func TestPointerReceiverCustomEncoderInErrorInterface(t *testing.T) {
	got, err := msgpack.Marshal([]error{pointerEncoderError("E42")})
	if err != nil {
		t.Fatal(err)
	}
	want, err := msgpack.Marshal([]interface{}{"custom:E42"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Marshal([]error{value}) = %x, want custom encoding %x", got, want)
	}
}
