package msgpack

import "testing"

// These tests live in the msgpack package (not msgpack_test) because they
// inspect the unexported wbuf/buf fields and the unexported retention
// predicate.
//
// Retention behavior is asserted through poolBufOversized rather than
// through Get/Put round-trips: sync.Pool may drop a Put item at any time
// (GC, or a Get on a different P), so "the big buffer came back" is not a
// property a test can rely on. The Get/Put tests below assert only the
// direction that stays true regardless of pool behavior — an oversized
// buffer is never handed back.

func TestPoolBufOversized(t *testing.T) {
	const big = 100 * 1024

	tests := []struct {
		name  string
		limit int // 0 means "leave at default"
		cap   int
		want  bool
	}{
		{name: "default/under", cap: 16 * 1024, want: false},
		{name: "default/at", cap: defaultPoolBufferLimit, want: false},
		{name: "default/over", cap: defaultPoolBufferLimit + 1, want: true},
		{name: "default/big", cap: big, want: true},

		{name: "raised/under", limit: 256 * 1024, cap: big, want: false},
		{name: "raised/at", limit: 256 * 1024, cap: 256 * 1024, want: false},
		{name: "raised/over", limit: 256 * 1024, cap: 256*1024 + 1, want: true},

		// A limit below the default clamps up, so the default still governs.
		{name: "clamped/under", limit: 1024, cap: 16 * 1024, want: false},
		{name: "clamped/over", limit: 1024, cap: big, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.limit != 0 {
				SetPoolBufferLimit(tt.limit)
				defer SetPoolBufferLimit(0)
			}
			if got := poolBufOversized(tt.cap); got != tt.want {
				t.Fatalf("poolBufOversized(%d) with limit %d = %v, want %v",
					tt.cap, getPoolBufferLimit(), got, tt.want)
			}
		})
	}
}

func TestPoolBufferLimitEncoder(t *testing.T) {
	const big = 100 * 1024

	// Under the default limit an oversized wbuf must never be handed back.
	// (A fresh encoder from the pool is also fine — both satisfy this.)
	enc := GetEncoder()
	enc.wbuf = make([]byte, big)
	PutEncoder(enc)
	enc = GetEncoder()
	if cap(enc.wbuf) > defaultPoolBufferLimit {
		t.Fatalf("wbuf cap=%d retained above default limit %d", cap(enc.wbuf), defaultPoolBufferLimit)
	}
	PutEncoder(enc)
}

func TestPoolBufferLimitDecoder(t *testing.T) {
	const big = 100 * 1024

	dec := GetDecoder()
	dec.buf = make([]byte, big)
	PutDecoder(dec)
	dec = GetDecoder()
	if cap(dec.buf) > defaultPoolBufferLimit {
		t.Fatalf("buf cap=%d retained above default limit %d", cap(dec.buf), defaultPoolBufferLimit)
	}
	PutDecoder(dec)
}

func TestSetPoolBufferLimitClamp(t *testing.T) {
	SetPoolBufferLimit(-1)
	if got := getPoolBufferLimit(); got != defaultPoolBufferLimit {
		t.Fatalf("limit=%d after SetPoolBufferLimit(-1), want default %d", got, defaultPoolBufferLimit)
	}
	SetPoolBufferLimit(64 * 1024)
	if got := getPoolBufferLimit(); got != 64*1024 {
		t.Fatalf("limit=%d, want %d", got, 64*1024)
	}
	// Values below the default are clamped up to the default.
	SetPoolBufferLimit(1024)
	if got := getPoolBufferLimit(); got != defaultPoolBufferLimit {
		t.Fatalf("limit=%d after SetPoolBufferLimit(1024), want clamped default %d", got, defaultPoolBufferLimit)
	}
	SetPoolBufferLimit(0)
	if got := getPoolBufferLimit(); got != defaultPoolBufferLimit {
		t.Fatalf("limit=%d after reset, want default %d", got, defaultPoolBufferLimit)
	}
}
