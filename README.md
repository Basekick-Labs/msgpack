# MessagePack encoding for Golang

[![Build Status](https://github.com/Basekick-Labs/msgpack/actions/workflows/build.yml/badge.svg?branch=v6)](https://github.com/Basekick-Labs/msgpack/actions/workflows/build.yml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Basekick-Labs/msgpack/v6)](https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6)
[![Discord](https://img.shields.io/badge/discord-chat-5865F2?logo=discord&logoColor=white)](https://discord.gg/nxnWfUxsdm)

> A performance-optimized fork of [vmihailenco/msgpack/v5](https://github.com/vmihailenco/msgpack),
> maintained by [Basekick Labs](https://github.com/Basekick-Labs). Built for
> [Arc](https://github.com/Basekick-Labs/arc), a high-performance time-series database.

## What's New in v6

**15 performance optimizations** — +8.6% throughput in Arc production benchmarks:
- Zero-allocation byte-slice reader for `Unmarshal()` (~21% faster, ~50% less memory)
- `*interface{}` fast path bypasses reflect for the most common decode pattern
- Pooled byte buffer and `MarshalAppend` API for zero-alloc encode
- Type-switch fast paths for `map[string]interface{}`, `map[string]string`, `[]interface{}`
- Two-pass `OmitEmpty`, `isZeroer` caching, `reflect.Convert` skip, sorted key pooling
- `sync.Pool` replacement for goroutine-per-type value cache

**15 bug fixes** — including OOM protection, float-to-int decoding, non-addressable pointer encode, and more.

See [CHANGELOG.md](CHANGELOG.md) for full details and the [v6.0.0 release](https://github.com/Basekick-Labs/msgpack/releases/tag/v6.0.0).

## Resources

- [Discord](https://discord.gg/nxnWfUxsdm)
- [Reference](https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6)
- [Examples](https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6#pkg-examples)

## Features

- Primitives, arrays, maps, structs, time.Time and interface{}.
- Appengine \*datastore.Key and datastore.Cursor.
- [CustomEncoder]/[CustomDecoder] interfaces for custom encoding.
- [Extensions](https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6#example-RegisterExt) to encode
  type information.
- Renaming fields via `msgpack:"my_field_name"` and alias via `msgpack:"alias:another_name"`.
- Omitting individual empty fields via `msgpack:",omitempty"` tag or all
  [empty fields in a struct](https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6#example-Marshal-OmitEmpty).
- [Map keys sorting](https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6#Encoder.SetSortMapKeys).
- Encoding/decoding all
  [structs as arrays](https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6#Encoder.UseArrayEncodedStructs)
  or
  [individual structs](https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6#example-Marshal-AsArray).
- [Encoder.SetCustomStructTag] with [Decoder.SetCustomStructTag] can turn msgpack into drop-in
  replacement for any tag.
- Simple but very fast and efficient
  [queries](https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6#example-Decoder.Query).

[customencoder]: https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6#CustomEncoder
[customdecoder]: https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6#CustomDecoder
[encoder.setcustomstructtag]:
  https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6#Encoder.SetCustomStructTag
[decoder.setcustomstructtag]:
  https://pkg.go.dev/github.com/Basekick-Labs/msgpack/v6#Decoder.SetCustomStructTag

## Installation

msgpack supports 2 last Go versions and requires support for
[Go modules](https://github.com/golang/go/wiki/Modules). So make sure to initialize a Go module:

```shell
go mod init github.com/my/repo
```

And then install msgpack:

```shell
go get github.com/Basekick-Labs/msgpack/v6
```

## Quickstart

```go
import "github.com/Basekick-Labs/msgpack/v6"

func ExampleMarshal() {
    type Item struct {
        Foo string
    }

    b, err := msgpack.Marshal(&Item{Foo: "bar"})
    if err != nil {
        panic(err)
    }

    var item Item
    err = msgpack.Unmarshal(b, &item)
    if err != nil {
        panic(err)
    }
    fmt.Println(item.Foo)
    // Output: bar
}
```

## Contributors

Thanks to all the people who already contributed!

<a href="https://github.com/Basekick-Labs/msgpack/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=Basekick-Labs/msgpack" />
</a>
