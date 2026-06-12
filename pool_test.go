package msgpack

import "testing"

// These tests live in the msgpack package (not msgpack_test) because they
// inspect the unexported wbuf/buf fields to verify pool retention behavior.
// They rely on sync.Pool returning the just-Put item on the same goroutine,
// which holds absent a GC between Put and Get.

func TestPoolBufferLimitEncoder(t *testing.T) {
	const big = 100 * 1024

	// Default limit: an oversized wbuf is dropped on Put.
	enc := GetEncoder()
	enc.wbuf = make([]byte, big)
	PutEncoder(enc)
	enc = GetEncoder()
	if cap(enc.wbuf) > defaultPoolBufferLimit {
		t.Fatalf("wbuf cap=%d retained above default limit", cap(enc.wbuf))
	}
	PutEncoder(enc)

	// Raised limit: the same buffer is retained.
	SetPoolBufferLimit(256 * 1024)
	defer SetPoolBufferLimit(0)

	enc = GetEncoder()
	enc.wbuf = make([]byte, big)
	PutEncoder(enc)
	enc = GetEncoder()
	if cap(enc.wbuf) < big {
		t.Fatalf("wbuf cap=%d, want >= %d retained under raised limit", cap(enc.wbuf), big)
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
		t.Fatalf("buf cap=%d retained above default limit", cap(dec.buf))
	}
	PutDecoder(dec)

	SetPoolBufferLimit(256 * 1024)
	defer SetPoolBufferLimit(0)

	dec = GetDecoder()
	dec.buf = make([]byte, big)
	PutDecoder(dec)
	dec = GetDecoder()
	if cap(dec.buf) < big {
		t.Fatalf("buf cap=%d, want >= %d retained under raised limit", cap(dec.buf), big)
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
