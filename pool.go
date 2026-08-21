package msgpack

import "sync/atomic"

// defaultPoolBufferLimit is the default maximum capacity, in bytes, of
// internal buffers retained by pooled encoders and decoders.
const defaultPoolBufferLimit = 32 * 1024

// poolBufferLimit holds the configured limit. Zero means "use the default";
// it is read atomically so SetPoolBufferLimit is safe to call concurrently
// with PutEncoder/PutDecoder.
var poolBufferLimit atomic.Int64

// SetPoolBufferLimit sets the maximum capacity, in bytes, of internal
// buffers retained by pooled encoders and decoders (GetEncoder/PutEncoder,
// GetDecoder/PutDecoder, and the package-level Marshal/MarshalAppend/
// Unmarshal helpers that use them). Buffers whose capacity exceeds the
// limit are dropped when the encoder or decoder is returned to the pool,
// so one-off large operations don't pin memory.
//
// The default is 32 KiB. Workloads that consistently encode or decode
// larger payloads can raise the limit to avoid re-growing buffers on every
// pooled use. Values below the default are clamped to the default;
// n <= 0 restores the default.
func SetPoolBufferLimit(n int) {
	if n <= 0 {
		poolBufferLimit.Store(0)
		return
	}
	if n < defaultPoolBufferLimit {
		n = defaultPoolBufferLimit
	}
	poolBufferLimit.Store(int64(n))
}

func getPoolBufferLimit() int {
	if n := poolBufferLimit.Load(); n > 0 {
		return int(n)
	}
	return defaultPoolBufferLimit
}

// poolBufOversized reports whether a pooled buffer of capacity c should be
// dropped rather than retained. The constant default check comes first so
// the common small-buffer case never pays for the atomic load; the
// configured limit is never below the default.
func poolBufOversized(c int) bool {
	return c > defaultPoolBufferLimit && c > getPoolBufferLimit()
}
