## Unreleased

### Performance

- **decode:** reuse caller-supplied destination map for `map[string]interface{}` — `Decode(&m)`/`Unmarshal(data, &m)` with a non-nil `m` now decodes into the existing map (entries merged) instead of replacing it with a fresh allocation, matching the long-standing `map[string]string` behavior. Callers that `clear(m)` and reuse the destination get zero map allocations per decode ([#61](https://github.com/Basekick-Labs/msgpack/issues/61)) (reused-destination decode **-22.9% ns/op**, **-80.8% B/op**). Note: this diverges from upstream, which replaces a non-nil `map[string]interface{}` destination; pass a nil map to keep replace semantics.

---

## v6.1.0 (2026-04-27)

### Performance

- **decode:** `readCode` bsr fast path — when decoding from a byte slice, reads directly from the underlying array instead of dispatching through the `io.ByteReader` interface; eliminates ~900M interface calls/sec at Arc's throughput ([#57](https://github.com/Basekick-Labs/msgpack/issues/57)) (StructUnmarshal **-7.5%**, StructUnmarshalPartially **-6.1%**)
- **decode:** `PeekCode` bsr fast path — peeks directly at `bsr.data[bsr.pos]` instead of `ReadByte` + `UnreadByte` (two interface calls) ([#59](https://github.com/Basekick-Labs/msgpack/issues/59))
- **encode:** pool `OmitEmpty` filtered field slices via `sync.Pool` — when fields are actually omitted, the allocated `[]*field` slice is now returned to a pool for reuse instead of being GC'd ([#58](https://github.com/Basekick-Labs/msgpack/issues/58))
- **encode/decode:** pool and pre-allocate interned-string dict — `SetInternedStringsDictCap(n)` pre-sizes the dict to avoid map rehashing and slice growth; pooled encoders/decoders now reuse dict storage across `Reset()` (cleared in place) instead of discarding it, and `Put*()` drops oversized dicts to keep the pool lean ([#66](https://github.com/Basekick-Labs/msgpack/issues/66))
- **decode:** hoist `newValue()` allocations out of `decodeTypedMapValue` loop — reuses a single key slot and value slot across all map entries, zeroing between iterations. Takes typed-map decode from 2N `reflect.New()` calls to 2 per map ([#65](https://github.com/Basekick-Labs/msgpack/issues/65)) (BenchmarkLargeMapIntInt **-50% allocs/op**, **-50% B/op**, **-10% ns/op** for 1000-entry `map[int]int`)

### Chores

- Lower `go.mod` directive from 1.26 to 1.25 — preserves drop-in compatibility for downstream users on Go 1.25; CI matrix unchanged (1.25.x, 1.26.x) ([#70](https://github.com/Basekick-Labs/msgpack/issues/70))

---

## v6.0.0 (Basekick-Labs fork)

### Performance

- **decode:** zero-allocation byte-slice reader for `Unmarshal()` — replaces `bytes.NewReader` + `bufio.NewReader` with a direct `byteSliceReader`, eliminating 2 allocations per call (~21% faster decode, ~50% less memory)
- **decode:** `*interface{}` fast path in `Decode()` — skips `reflect.ValueOf` for the most common `Unmarshal(b, &interface{})` pattern (~14% faster)
- **encode:** pooled byte buffer in `Marshal()` — replaces per-call `bytes.Buffer` with a reusable `[]byte` embedded in the pooled `Encoder` struct
- **encode:** `byteSliceWriter` for `Marshal()` path — native `WriteByte` implementation eliminates per-byte heap allocation
- **encode:** `byteWriter.WriteByte` scratch fix for streaming path — uses `[1]byte` scratch instead of allocating `[]byte{c}`
- **encode:** `Encode()` fast paths for `map[string]interface{}` and `[]interface{}` — bypasses `reflect.ValueOf` + sync.Map encoder lookup
- **encode:** `MarshalAppend(dst, v)` API — appends encoded bytes to caller-provided buffer, eliminating the final `make+copy` in `Marshal()` (-26% faster, -94% less memory for callers who reuse buffers)
- **encode:** two-pass `OmitEmpty` — avoids slice allocation when no fields are omitted (common case for time-series data)
- **encode:** cache `isZeroer` interface check — pre-computes at struct-discovery time to skip `v.Interface()` boxing during `OmitEmpty` checks
- **encode/decode:** skip `reflect.Convert` for exact map/slice types — adds `v.Type() == targetType` fast path before `Convert()` for `map[string]string`, `map[string]bool`, `map[string]interface{}`, and `[]string` (-7–8% faster)
- **encode:** `map[string]string` fast path in `Encode()` type switch — bypasses `reflect.ValueOf` + encoder lookup (-15% faster)
- **decode:** replace goroutine-per-type `cachedValues` with `sync.Pool` — eliminates goroutine leak and channel synchronization overhead
- **encode:** pool sorted map key slices via `sync.Pool` — eliminates 1 alloc per sorted map encode for `SetSortMapKeys(true)` callers
- **decode:** pool recording buffer in `unmarshalValue` — eliminates 1 alloc per `Unmarshaler.UnmarshalMsgpack` call
- **decode:** inline `hasNilCode` for byte-slice reader path — peeks directly at underlying data, avoiding two interface method calls per nil check (-2–4% faster decode)

### Bug Fixes

- **decode:** cap `decodeSlice()` allocation at `sliceAllocLimit` (1M) to prevent OOM from malicious payloads ([#1](https://github.com/Basekick-Labs/msgpack/issues/1))
- **decode:** cap `DecodeMap()` allocation at `maxMapSize` (1M) — same OOM vector for `map[string]interface{}` path
- **decode:** fix `disableAllocLimitFlag` check in `decodeSliceValue` — `!= 1` was always true because the flag value is `1 << 3 = 8`, so the alloc limit in `growSliceValue()` was never applied
- **decode:** fix error message in `DecodeFloat64` — said "decoding float32" instead of "decoding float64" ([#13](https://github.com/Basekick-Labs/msgpack/issues/13))
- **decode:** allow float-encoded values to decode into `int64`/`uint64` with validation — rejects NaN, Inf, fractional, and out-of-range values ([#2](https://github.com/Basekick-Labs/msgpack/issues/2))
- **decode:** allow `float64`-encoded values to decode into `float32` with overflow check ([#12](https://github.com/Basekick-Labs/msgpack/issues/12))
- **encode:** handle non-addressable values with pointer receivers — `ensureAddr` creates an addressable copy instead of returning an error ([#3](https://github.com/Basekick-Labs/msgpack/issues/3))
- **encode:** prevent panic when marshalling `reflect.Value` — unwraps and encodes the underlying value ([#15](https://github.com/Basekick-Labs/msgpack/issues/15))
- **encode:** preserve custom error types instead of reducing to plain strings via `.Error()` ([#22](https://github.com/Basekick-Labs/msgpack/issues/22))
- **decode:** support non-string map keys when decoding into `interface{}` ([#21](https://github.com/Basekick-Labs/msgpack/issues/21))
- **decode:** handle unaddressable values in interface decode ([#21](https://github.com/Basekick-Labs/msgpack/issues/21))
- **encode:** respect non-zero unexported fields in `omitempty` struct emptiness check ([#6](https://github.com/Basekick-Labs/msgpack/issues/6))
- **decode:** use `interface{}` value type for non-string-keyed typed maps to support heterogeneous nested maps ([#20](https://github.com/Basekick-Labs/msgpack/issues/20))
- **pool:** drop oversized decoder buffers (>32KB) to prevent memory leak from large decode operations ([#19](https://github.com/Basekick-Labs/msgpack/issues/19))
- **decode:** choose `TextUnmarshaler` over `BinaryUnmarshaler` when wire format is `str` ([#10](https://github.com/Basekick-Labs/msgpack/issues/10))

### Chores

- Modernize GitHub Actions (checkout@v4, setup-go@v5)
- Go version matrix: 1.25.x, 1.26.x
- Bump `go.mod` to Go 1.26
- CI: add `-count=1 -timeout=5m` and `GOGC=50` to race tests to prevent OOM on runners ([#33](https://github.com/Basekick-Labs/msgpack/issues/33))
- CI: change cross-platform step from `go test` to `go vet` (compile-only)
- Remove tautological `if err != nil` in `decodeInternedInterfaceValue` (nilness warning)
- Remove unused `nilableType` function from `encode_value.go`

---

## [5.4.1](https://github.com/vmihailenco/msgpack/compare/v5.4.0...v5.4.1) (2023-10-26)


### Bug Fixes

* **reflect:** not assignable to type ([edeaedd](https://github.com/vmihailenco/msgpack/commit/edeaeddb2d51868df8c6ff2d8a218b527aeaf5fd))



# [5.4.0](https://github.com/vmihailenco/msgpack/compare/v5.3.6...v5.4.0) (2023-10-01)



## [5.3.6](https://github.com/vmihailenco/msgpack/compare/v5.3.5...v5.3.6) (2023-10-01)


### Features

* allow overwriting time.Time parsing from extID 13 (for NodeJS Date) ([9a6b73b](https://github.com/vmihailenco/msgpack/commit/9a6b73b3588fd962d568715f4375e24b089f7066))
* apply omitEmptyFlag to empty structs ([e5f8d03](https://github.com/vmihailenco/msgpack/commit/e5f8d03c0a1dd9cc571d648cd610305139078de5))
* support sorted keys for map[string]bool ([690c1fa](https://github.com/vmihailenco/msgpack/commit/690c1fab9814fab4842295ea986111f49850d9a4))



## [5.3.5](https://github.com/vmihailenco/msgpack/compare/v5.3.4...v5.3.5) (2021-10-22)

- Allow decoding `nil` code as boolean false.

## v5

### Added

- `DecodeMap` is split into `DecodeMap`, `DecodeTypedMap`, and `DecodeUntypedMap`.
- New msgpack extensions API.

### Changed

- `Reset*` functions also reset flags.
- `SetMapDecodeFunc` is renamed to `SetMapDecoder`.
- `StructAsArray` is renamed to `UseArrayEncodedStructs`.
- `SortMapKeys` is renamed to `SetSortMapKeys`.

### Removed

- `UseJSONTag` is removed. Use `SetCustomStructTag("json")` instead.

## v4

- Encode, Decode, Marshal, and Unmarshal are changed to accept single argument. EncodeMulti and
  DecodeMulti are added as replacement.
- Added EncodeInt8/16/32/64 and EncodeUint8/16/32/64.
- Encoder changed to preserve type of numbers instead of chosing most compact encoding. The old
  behavior can be achieved with Encoder.UseCompactEncoding.

## v3.3

- `msgpack:",inline"` tag is restored to force inlining structs.

## v3.2

- Decoding extension types returns pointer to the value instead of the value. Fixes #153

## v3

- gopkg.in is not supported any more. Update import path to github.com/vmihailenco/msgpack.
- Msgpack maps are decoded into map[string]interface{} by default.
- EncodeSliceLen is removed in favor of EncodeArrayLen. DecodeSliceLen is removed in favor of
  DecodeArrayLen.
- Embedded structs are automatically inlined where possible.
- Time is encoded using extension as described in https://github.com/msgpack/msgpack/pull/209. Old
  format is supported as well.
- EncodeInt8/16/32/64 is replaced with EncodeInt. EncodeUint8/16/32/64 is replaced with EncodeUint.
  There should be no performance differences.
- DecodeInterface can now return int8/16/32 and uint8/16/32.
- PeekCode returns codes.Code instead of byte.
