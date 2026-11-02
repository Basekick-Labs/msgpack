package msgpack_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/Basekick-Labs/msgpack/v6"
	"github.com/stretchr/testify/require"
)

type textBytes []byte

func (b textBytes) MarshalText() ([]byte, error)     { return []byte("text"), nil }
func (b *textBytes) UnmarshalText(data []byte) error { *b = append((*b)[:0], data...); return nil }

type textByteArray [2]byte

func (b *textByteArray) MarshalText() ([]byte, error)    { return []byte("text"), nil }
func (b *textByteArray) UnmarshalText(data []byte) error { copy(b[:], data); return nil }

func TestTextOptOutByteValues(t *testing.T) {
	slice := textBytes{1, 2}
	for _, tc := range []struct {
		name           string
		in, out, plain interface{}
	}{
		{"slice", textBytes{1, 2}, new(textBytes), []byte{1, 2}},
		{"slice-pointer", &slice, new(textBytes), []byte{1, 2}},
		{"array", textByteArray{1, 2}, new(textByteArray), [2]byte{1, 2}},
		{"array-pointer", &textByteArray{1, 2}, new(textByteArray), [2]byte{1, 2}},
		{"nil-slice", textBytes(nil), new(textBytes), []byte(nil)},
		{"empty-slice", textBytes{}, new(textBytes), []byte{}},
		{"addressable-array-field", &struct{ Value textByteArray }{textByteArray{1, 2}}, new(struct{ Value textByteArray }), struct{ Value [2]byte }{[2]byte{1, 2}}},
		{"addressable-slice-field", &struct{ Value textBytes }{textBytes{1, 2}}, new(struct{ Value textBytes }), struct{ Value []byte }{[]byte{1, 2}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expected, err := msgpack.Marshal(tc.plain)
			require.NoError(t, err)
			t.Run("encode", func(t *testing.T) {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				enc.UseTextMarshaler(false)
				require.NoError(t, enc.Encode(tc.in))
				require.Equal(t, expected, buf.Bytes())
			})
			t.Run("decode", func(t *testing.T) {
				dec := msgpack.NewDecoder(bytes.NewReader(expected))
				dec.UseTextUnmarshaler(false)
				require.NoError(t, dec.Decode(tc.out))
				require.Equal(t, reflect.Indirect(reflect.ValueOf(tc.in)).Interface(), reflect.Indirect(reflect.ValueOf(tc.out)).Interface())
			})
		})
	}
}
