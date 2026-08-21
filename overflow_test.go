package msgpack

import (
	"math"
	"strings"
	"testing"
)

// uint32Len is the single conversion point for 32-bit msgpack length
// headers. On 64-bit platforms every uint32 fits in an int; on 32-bit
// platforms values above math.MaxInt32 must be rejected rather than
// wrapped to a negative int (which callers would read as nil, or pass to
// make()). The table drives uint32Len directly so both behaviors are
// covered regardless of the host word size.
func TestUint32LenOverflow(t *testing.T) {
	tests := []struct {
		name    string
		n       uint32
		wantErr bool
	}{
		{name: "zero", n: 0},
		{name: "small", n: 42},
		{name: "maxint32", n: math.MaxInt32},
		// Wraps to a negative int on 32-bit; fine on 64-bit.
		{name: "maxint32+1", n: math.MaxInt32 + 1, wantErr: intIs32Bit},
		// Wraps to exactly -1 on 32-bit -- the "silently decodes as nil" case.
		{name: "maxuint32", n: math.MaxUint32, wantErr: intIs32Bit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := uint32Len(tt.n, nil, "test")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("uint32Len(%d) = %d, want overflow error", tt.n, got)
				}
				if !strings.Contains(err.Error(), "overflows int") {
					t.Fatalf("uint32Len(%d) err = %v, want overflow error", tt.n, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("uint32Len(%d) unexpected error: %v", tt.n, err)
			}
			if got < 0 {
				t.Fatalf("uint32Len(%d) = %d, must never be negative", tt.n, got)
			}
			if uint32(got) != tt.n {
				t.Fatalf("uint32Len(%d) = %d, want %d", tt.n, got, tt.n)
			}
		})
	}
}

const intIs32Bit = math.MaxInt == math.MaxInt32

// An existing error is passed through untouched.
func TestUint32LenPropagatesError(t *testing.T) {
	sentinel := errTestSentinel
	if _, err := uint32Len(math.MaxUint32, sentinel, "test"); err != sentinel {
		t.Fatalf("uint32Len err = %v, want the incoming error", err)
	}
}

var errTestSentinel = errSentinel{}

type errSentinel struct{}

func (errSentinel) Error() string { return "sentinel" }

// Negative lengths must never reach make() or a slice expression.
func TestReadNRejectsNegative(t *testing.T) {
	d := NewDecoder(strings.NewReader("whatever"))
	if _, err := d.readN(-1); err == nil {
		t.Fatal("d.readN(-1) = nil error, want invalid length error")
	}
	if _, err := readN(strings.NewReader("x"), nil, -1); err == nil {
		t.Fatal("readN(-1) = nil error, want invalid length error")
	}
	if _, err := readNGrow(strings.NewReader("x"), nil, -1); err == nil {
		t.Fatal("readNGrow(-1) = nil error, want invalid length error")
	}
}

// Forces the 32-bit rejection path on any host by lowering the ceiling to
// math.MaxInt32, so CI proves the guard works without a 32-bit runner.
func TestUint32LenRejectsOverflowAs32Bit(t *testing.T) {
	orig := maxLenForInt
	maxLenForInt = math.MaxInt32
	defer func() { maxLenForInt = orig }()

	for _, n := range []uint32{math.MaxInt32 + 1, 0x80000000, math.MaxUint32} {
		got, err := uint32Len(n, nil, "test")
		if err == nil {
			t.Fatalf("uint32Len(%#x) = %d, want overflow error when int is 32-bit", n, got)
		}
	}
	// Values that still fit must keep working.
	if got, err := uint32Len(math.MaxInt32, nil, "test"); err != nil || got != math.MaxInt32 {
		t.Fatalf("uint32Len(MaxInt32) = %d, %v; want %d, nil", got, err, int64(math.MaxInt32))
	}
}

// End-to-end: malicious 32-bit headers decoded through the public API with
// the ceiling forced to 32-bit. Each must return an error rather than
// panicking in make() or silently decoding as nil.
func TestDecodeRejectsOverflowingHeadersAs32Bit(t *testing.T) {
	orig := maxLenForInt
	maxLenForInt = math.MaxInt32
	defer func() { maxLenForInt = orig }()

	tests := []struct {
		name string
		data []byte
		dst  func() interface{}
	}{
		{
			name: "map32 maxuint32",
			data: []byte{0xdf, 0xff, 0xff, 0xff, 0xff},
			dst:  func() interface{} { return &map[string]interface{}{} },
		},
		{
			name: "array32 maxuint32",
			data: []byte{0xdd, 0xff, 0xff, 0xff, 0xff},
			dst:  func() interface{} { return &[]interface{}{} },
		},
		{
			name: "str32 maxuint32",
			data: []byte{0xdb, 0xff, 0xff, 0xff, 0xff},
			dst:  func() interface{} { var s string; return &s },
		},
		{
			name: "bin32 maxuint32",
			data: []byte{0xc6, 0xff, 0xff, 0xff, 0xff},
			dst:  func() interface{} { var b []byte; return &b },
		},
		{
			name: "str32 0x80000000",
			data: []byte{0xdb, 0x80, 0x00, 0x00, 0x00},
			dst:  func() interface{} { var s string; return &s },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic decoding %s: %v", tt.name, r)
				}
			}()
			err := Unmarshal(tt.data, tt.dst())
			if err == nil {
				t.Fatalf("Unmarshal(%s) = nil error, want overflow rejection", tt.name)
			}
			// Must be rejected at the length check, not incidentally by a
			// later EOF -- that distinction is the whole point of the guard.
			if !strings.Contains(err.Error(), "overflows int") {
				t.Fatalf("Unmarshal(%s) err = %v, want an \"overflows int\" rejection", tt.name, err)
			}
		})
	}
}
