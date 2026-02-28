## v6 (Basekick-Labs fork)

### Performance

- **decode:** zero-allocation byte-slice reader for `Unmarshal()` — replaces `bytes.NewReader` + `bufio.NewReader` with a direct `byteSliceReader`, eliminating 2 allocations per call (~21% faster decode, ~50% less memory)
- **decode:** `*interface{}` fast path in `Decode()` — skips `reflect.ValueOf` for the most common `Unmarshal(b, &interface{})` pattern (~14% faster)
- **encode:** pooled byte buffer in `Marshal()` — replaces per-call `bytes.Buffer` with a reusable `[]byte` embedded in the pooled `Encoder` struct
- **encode:** `byteSliceWriter` for `Marshal()` path — native `WriteByte` implementation eliminates per-byte heap allocation
- **encode:** `byteWriter.WriteByte` scratch fix for streaming path — uses `[1]byte` scratch instead of allocating `[]byte{c}`
- **encode:** `Encode()` fast paths for `map[string]interface{}` and `[]interface{}` — bypasses `reflect.ValueOf` + sync.Map encoder lookup

### Bug Fixes

- **decode:** cap `decodeSlice()` allocation at `sliceAllocLimit` (1M) to prevent OOM from malicious payloads ([#1](https://github.com/Basekick-Labs/msgpack/issues/1))
- **decode:** cap `DecodeMap()` allocation at `maxMapSize` (1M) — same OOM vector for `map[string]interface{}` path
- **decode:** fix `disableAllocLimitFlag` check in `decodeSliceValue` — `!= 1` was always true because the flag value is `1 << 3 = 8`, so the alloc limit in `growSliceValue()` was never applied
- **decode:** fix error message in `DecodeFloat64` — said "decoding float32" instead of "decoding float64" ([#13](https://github.com/Basekick-Labs/msgpack/issues/13))

### Chores

- Modernize GitHub Actions (checkout@v4, setup-go@v5)
- Go version matrix: 1.25.x, 1.26.x
- Bump `go.mod` to Go 1.26

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
