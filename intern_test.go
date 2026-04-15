package msgpack_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/Basekick-Labs/msgpack/v6"
)

type NoIntern struct {
	A string
	B string
	C interface{}
}

type Intern struct {
	A string      `msgpack:",intern"`
	B string      `msgpack:",intern"`
	C interface{} `msgpack:",intern"`
}

func TestInternedString(t *testing.T) {
	var buf bytes.Buffer

	enc := msgpack.NewEncoder(&buf)
	enc.UseInternedStrings(true)

	dec := msgpack.NewDecoder(&buf)
	dec.UseInternedStrings(true)

	for i := 0; i < 3; i++ {
		err := enc.EncodeString("hello")
		require.Nil(t, err)
	}

	for i := 0; i < 3; i++ {
		s, err := dec.DecodeString()
		require.Nil(t, err)
		require.Equal(t, "hello", s)
	}

	err := enc.Encode("hello")
	require.Nil(t, err)

	v, err := dec.DecodeInterface()
	require.Nil(t, err)
	require.Equal(t, "hello", v)

	_, err = dec.DecodeInterface()
	require.Equal(t, io.EOF, err)
}

func TestInternedStringTag(t *testing.T) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	dec := msgpack.NewDecoder(&buf)

	in := []Intern{
		{"f", "f", "f"},
		{"fo", "fo", "fo"},
		{"foo", "foo", "foo"},
		{"f", "fo", "foo"},
	}
	err := enc.Encode(in)
	require.Nil(t, err)

	var out []Intern
	err = dec.Decode(&out)
	require.Nil(t, err)
	require.Equal(t, in, out)
}

func TestResetDict(t *testing.T) {
	dict := []string{"hello world", "foo bar"}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	dec := msgpack.NewDecoder(&buf)

	{
		enc.ResetDict(&buf, dictMap(dict))
		err := enc.EncodeString("hello world")
		require.Nil(t, err)
		require.Equal(t, 3, buf.Len())

		dec.ResetDict(&buf, dict)
		s, err := dec.DecodeString()
		require.Nil(t, err)
		require.Equal(t, "hello world", s)
	}

	{
		enc.ResetDict(&buf, dictMap(dict))
		err := enc.Encode("foo bar")
		require.Nil(t, err)
		require.Equal(t, 3, buf.Len())

		dec.ResetDict(&buf, dict)
		s, err := dec.DecodeInterface()
		require.Nil(t, err)
		require.Equal(t, "foo bar", s)
	}

	dec.ResetDict(&buf, dict)
	_ = enc.EncodeString("xxxx")
	require.Equal(t, 5, buf.Len())
	_ = enc.Encode("xxxx")
	require.Equal(t, 10, buf.Len())
}

func TestMapWithInternedString(t *testing.T) {
	type M map[string]interface{}

	dict := []string{"hello world", "foo bar"}

	var buf bytes.Buffer

	enc := msgpack.NewEncoder(nil)
	enc.ResetDict(&buf, dictMap(dict))

	dec := msgpack.NewDecoder(nil)
	dec.ResetDict(&buf, dict)

	for i := 0; i < 100; i++ {
		in := M{
			"foo bar":     "hello world",
			"hello world": "foo bar",
			"foo":         "bar",
		}
		err := enc.Encode(in)
		require.Nil(t, err)

		_, err = dec.DecodeInterface()
		require.Nil(t, err)
	}
}

func dictMap(dict []string) map[string]int {
	m := make(map[string]int, len(dict))
	for i, s := range dict {
		m[s] = i
	}
	return m
}

func TestInternedStringDictCap(t *testing.T) {
	var buf bytes.Buffer

	enc := msgpack.NewEncoder(&buf)
	enc.UseInternedStrings(true)
	enc.SetInternedStringsDictCap(1024)

	dec := msgpack.NewDecoder(&buf)
	dec.UseInternedStrings(true)
	dec.SetInternedStringsDictCap(1024)

	const n = 2000 // deliberately more than the cap hint; it is a hint, not a limit
	in := make([]string, n)
	for i := 0; i < n; i++ {
		in[i] = "key_" + string(rune('a'+i%26)) + "_" + string(rune('0'+i%10)) + "_" +
			string(rune('A'+(i/10)%26)) + "_" + string(rune('0'+(i/100)%10))
	}

	for _, s := range in {
		require.Nil(t, enc.EncodeString(s))
	}
	for i := 0; i < n; i++ {
		s, err := dec.DecodeString()
		require.Nil(t, err)
		require.Equal(t, in[i], s)
	}
}

func TestInternedStringDictReuseAcrossReset(t *testing.T) {
	var buf1, buf2 bytes.Buffer

	enc := msgpack.NewEncoder(&buf1)
	enc.UseInternedStrings(true)

	dec := msgpack.NewDecoder(&buf1)
	dec.UseInternedStrings(true)

	for i := 0; i < 5; i++ {
		require.Nil(t, enc.EncodeString("first-session"))
	}
	for i := 0; i < 5; i++ {
		s, err := dec.DecodeString()
		require.Nil(t, err)
		require.Equal(t, "first-session", s)
	}

	// Reset (not ResetDict) — dict storage should be cleared in place,
	// and a new session must not see prior-session entries.
	enc.Reset(&buf2)
	enc.UseInternedStrings(true)
	dec.Reset(&buf2)
	dec.UseInternedStrings(true)

	for i := 0; i < 5; i++ {
		require.Nil(t, enc.EncodeString("second-session"))
	}
	for i := 0; i < 5; i++ {
		s, err := dec.DecodeString()
		require.Nil(t, err)
		require.Equal(t, "second-session", s)
	}
}

// TestResetDoesNotMutateCallerDict guards against silently clearing or
// aliasing a caller-supplied dict via Reset after ResetDict. Regression
// for the ownership bug fixed alongside #66.
func TestResetDoesNotMutateCallerDict(t *testing.T) {
	externalEnc := map[string]int{"hello world": 0, "foo bar": 1}
	externalDec := []string{"hello world", "foo bar"}

	var buf bytes.Buffer

	enc := msgpack.NewEncoder(&buf)
	enc.ResetDict(&buf, externalEnc)
	enc.Reset(&buf) // must NOT clear externalEnc
	require.Equal(t, 2, len(externalEnc))
	require.Equal(t, 0, externalEnc["hello world"])
	require.Equal(t, 1, externalEnc["foo bar"])

	dec := msgpack.NewDecoder(&buf)
	dec.ResetDict(&buf, externalDec)
	dec.ResetBytes(nil) // must NOT alias externalDec's backing array

	// Now encode an interned string with a fresh encoder and decode it
	// through `dec`. If ResetBytes failed to drop the external alias, the
	// decoded string would be appended into externalDec.
	var payload bytes.Buffer
	encFresh := msgpack.NewEncoder(&payload)
	encFresh.UseInternedStrings(true)
	require.Nil(t, encFresh.EncodeString("would-clobber"))

	dec.ResetBytes(payload.Bytes())
	dec.UseInternedStrings(true)
	s, err := dec.DecodeString()
	require.Nil(t, err)
	require.Equal(t, "would-clobber", s)

	require.Equal(t, []string{"hello world", "foo bar"}, externalDec)
}

// TestInternedStringDictStorageIsReused verifies that the underlying map
// bucket array and slice backing array are kept across Reset — i.e., the
// clear-in-place / truncate-in-place path actually reuses storage and
// doesn't reallocate.
func TestInternedStringDictStorageIsReused(t *testing.T) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseInternedStrings(true)
	dec := msgpack.NewDecoder(&buf)
	dec.UseInternedStrings(true)

	// Warm both dicts so storage exists.
	require.Nil(t, enc.EncodeString("warmup-string"))
	_, err := dec.DecodeString()
	require.Nil(t, err)

	// Pre-size the buffer so bytes.Buffer growth doesn't pollute the count.
	buf.Grow(64)

	encAllocs := testing.AllocsPerRun(50, func() {
		buf.Reset()
		enc.Reset(&buf)
		enc.UseInternedStrings(true)
		_ = enc.EncodeString("warmup-string")
	})
	// With dict storage reuse, Reset + interned encode of an already-known
	// string should allocate nothing. Without reuse, the map is freshly
	// allocated each call.
	require.Equalf(t, float64(0), encAllocs,
		"encoder intern dict storage not reused: %v allocs/op", encAllocs)

	// For the decoder, drive it from a pre-built byte slice so we can
	// isolate dict allocations from reader setup. Encode the string twice:
	// the first is the raw interned-tagged string (which the decoder must
	// allocate a Go string for), the second is a 3-byte ext index reference
	// that decodes by fetching from the dict with zero allocations.
	var payload bytes.Buffer
	encFresh := msgpack.NewEncoder(&payload)
	encFresh.UseInternedStrings(true)
	require.Nil(t, encFresh.EncodeString("warmup-string"))
	require.Nil(t, encFresh.EncodeString("warmup-string"))
	data := payload.Bytes()

	decAllocs := testing.AllocsPerRun(50, func() {
		dec.ResetBytes(data)
		dec.UseInternedStrings(true)
		_, _ = dec.DecodeString() // populates dict[0]="warmup-string"
		_, _ = dec.DecodeString() // fetches dict[0] by index — no alloc
	})
	// The first DecodeString allocates one Go string for the raw payload
	// and appends it to the dict (no slice growth thanks to reuse); the
	// second DecodeString reads an ext index and returns the interned
	// string without allocating. Total: 1 alloc/op for the first string.
	require.Equalf(t, float64(1), decAllocs,
		"decoder intern dict storage not reused: %v allocs/op", decAllocs)
}

func TestInternedStringDictPoolRecycle(t *testing.T) {
	// First pooled session: intern some strings, then return to the pool.
	enc1 := msgpack.GetEncoder()
	var buf1 bytes.Buffer
	enc1.Reset(&buf1)
	enc1.UseInternedStrings(true)
	for i := 0; i < 3; i++ {
		require.Nil(t, enc1.EncodeString("leaked-if-buggy"))
	}
	msgpack.PutEncoder(enc1)

	// Second pooled session: may reuse enc1. Its dict must not leak the
	// prior session's entry — either it was cleared or replaced.
	enc2 := msgpack.GetEncoder()
	var buf2 bytes.Buffer
	enc2.Reset(&buf2)
	enc2.UseInternedStrings(true)
	require.Nil(t, enc2.EncodeString("fresh-session"))
	msgpack.PutEncoder(enc2)

	dec := msgpack.NewDecoder(&buf2)
	dec.UseInternedStrings(true)
	s, err := dec.DecodeString()
	require.Nil(t, err)
	require.Equal(t, "fresh-session", s)
}
