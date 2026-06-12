package msgpack_test

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/Basekick-Labs/msgpack/v6"
)

type nameStruct struct {
	Name string
}

type MsgpackTest struct {
	suite.Suite

	buf *bytes.Buffer
	enc *msgpack.Encoder
	dec *msgpack.Decoder
}

func (t *MsgpackTest) SetupTest() {
	t.buf = &bytes.Buffer{}
	t.enc = msgpack.NewEncoder(t.buf)
	t.dec = msgpack.NewDecoder(bufio.NewReader(t.buf))
}

func TestMsgpackTestSuite(t *testing.T) {
	suite.Run(t, new(MsgpackTest))
}

func (t *MsgpackTest) TestDecodeNil() {
	t.NotNil(t.dec.Decode(nil))
}

func (t *MsgpackTest) TestTime() {
	in := time.Now()
	var out time.Time

	t.Nil(t.enc.Encode(in))
	t.Nil(t.dec.Decode(&out))
	t.True(out.Equal(in))

	var zero time.Time
	t.Nil(t.enc.Encode(zero))
	t.Nil(t.dec.Decode(&out))
	t.True(out.Equal(zero))
	t.True(out.IsZero())
}

func (t *MsgpackTest) TestLargeBytes() {
	N := int(1e6)

	src := bytes.Repeat([]byte{'1'}, N)
	t.Nil(t.enc.Encode(src))
	var dst []byte
	t.Nil(t.dec.Decode(&dst))
	t.Equal(dst, src)
}

func (t *MsgpackTest) TestLargeString() {
	N := int(1e6)

	src := string(bytes.Repeat([]byte{'1'}, N))
	t.Nil(t.enc.Encode(src))
	var dst string
	t.Nil(t.dec.Decode(&dst))
	t.Equal(dst, src)
}

func (t *MsgpackTest) TestSliceOfStructs() {
	in := []*nameStruct{{"hello"}}
	var out []*nameStruct
	t.Nil(t.enc.Encode(in))
	t.Nil(t.dec.Decode(&out))
	t.Equal(out, in)
}

func (t *MsgpackTest) TestMap() {
	for _, i := range []struct {
		m map[string]string
		b []byte
	}{
		{map[string]string{}, []byte{0x80}},
		{map[string]string{"hello": "world"}, []byte{0x81, 0xa5, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0xa5, 0x77, 0x6f, 0x72, 0x6c, 0x64}},
	} {
		t.Nil(t.enc.Encode(i.m))
		t.Equal(t.buf.Bytes(), i.b, fmt.Errorf("err encoding %v", i.m))
		var m map[string]string
		t.Nil(t.dec.Decode(&m))
		t.Equal(m, i.m)
	}
}

func (t *MsgpackTest) TestMapStringInterfaceReuse() {
	in := map[string]interface{}{"hello": "world", "n": int8(1)}
	t.Nil(t.enc.Encode(in))

	// A caller-supplied map is decoded into in place (entries merged),
	// matching the map[string]string behavior.
	dst := map[string]interface{}{"existing": true, "hello": "stale"}
	t.Nil(t.dec.Decode(&dst))
	t.Equal(map[string]interface{}{
		"existing": true, // retained
		"hello":    "world",
		"n":        int8(1),
	}, dst)

	// msgpack nil still resets the destination to nil.
	t.Nil(t.enc.Encode(map[string]interface{}(nil)))
	t.Nil(t.dec.Decode(&dst))
	t.Nil(dst)

	// And a nil destination map is allocated as before.
	t.Nil(t.enc.Encode(in))
	dst = nil
	t.Nil(t.dec.Decode(&dst))
	t.Equal(in, dst)

	// An empty msgpack map leaves a pre-populated destination untouched.
	t.Nil(t.enc.Encode(map[string]interface{}{}))
	dst = map[string]interface{}{"existing": true}
	t.Nil(t.dec.Decode(&dst))
	t.Equal(map[string]interface{}{"existing": true}, dst)
}

func (t *MsgpackTest) TestMapStringInterfaceStructFieldReuse() {
	// Struct fields of type map[string]interface{} get the same merge
	// semantics as map[string]string fields always had: a pre-populated
	// field map is decoded into, not replaced.
	type withMap struct {
		Data map[string]interface{}
	}

	t.Nil(t.enc.Encode(withMap{Data: map[string]interface{}{"new": "value"}}))

	out := withMap{Data: map[string]interface{}{"existing": true}}
	t.Nil(t.dec.Decode(&out))
	t.Equal(map[string]interface{}{"existing": true, "new": "value"}, out.Data)
}

func (t *MsgpackTest) TestStructNil() {
	var dst *nameStruct

	t.Nil(t.enc.Encode(nameStruct{Name: "foo"}))
	t.Nil(t.dec.Decode(&dst))
	t.NotNil(dst)
	t.Equal(dst.Name, "foo")
}

func (t *MsgpackTest) TestStructUnknownField() {
	in := struct {
		Field1 string
		Field2 string
		Field3 string
	}{
		Field1: "value1",
		Field2: "value2",
		Field3: "value3",
	}
	t.Nil(t.enc.Encode(in))

	out := struct {
		Field2 string
	}{}
	t.Nil(t.dec.Decode(&out))
	t.Equal(out.Field2, "value2")
}

//------------------------------------------------------------------------------

type coderStruct struct {
	name string
}

type wrapperStruct struct {
	coderStruct
}

var (
	_ msgpack.CustomEncoder = (*coderStruct)(nil)
	_ msgpack.CustomDecoder = (*coderStruct)(nil)
)

func (s *coderStruct) Name() string {
	return s.name
}

func (s *coderStruct) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(s.name)
}

func (s *coderStruct) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.Decode(&s.name)
}

func (t *MsgpackTest) TestCoder() {
	in := &coderStruct{name: "hello"}
	var out coderStruct
	t.Nil(t.enc.Encode(in))
	t.Nil(t.dec.Decode(&out))
	t.Equal(out.Name(), "hello")
}

func (t *MsgpackTest) TestNilCoder() {
	in := &coderStruct{name: "hello"}
	var out *coderStruct
	t.Nil(t.enc.Encode(in))
	t.Nil(t.dec.Decode(&out))
	t.Equal(out.Name(), "hello")
}

func (t *MsgpackTest) TestNilCoderValue() {
	in := &coderStruct{name: "hello"}
	var out *coderStruct
	t.Nil(t.enc.Encode(in))
	t.Nil(t.dec.DecodeValue(reflect.ValueOf(&out)))
	t.Equal(out.Name(), "hello")
}

func (t *MsgpackTest) TestPtrToCoder() {
	in := &coderStruct{name: "hello"}
	var out coderStruct
	out2 := &out
	t.Nil(t.enc.Encode(in))
	t.Nil(t.dec.Decode(&out2))
	t.Equal(out.Name(), "hello")
}

func (t *MsgpackTest) TestWrappedCoder() {
	in := &wrapperStruct{coderStruct: coderStruct{name: "hello"}}
	var out wrapperStruct
	t.Nil(t.enc.Encode(in))
	t.Nil(t.dec.Decode(&out))
	t.Equal(out.Name(), "hello")
}

//------------------------------------------------------------------------------

type struct2 struct {
	Name string
}

type struct1 struct {
	Name    string
	Struct2 struct2
}

func (t *MsgpackTest) TestNestedStructs() {
	in := &struct1{Name: "hello", Struct2: struct2{Name: "world"}}
	var out struct1
	t.Nil(t.enc.Encode(in))
	t.Nil(t.dec.Decode(&out))
	t.Equal(out.Name, in.Name)
	t.Equal(out.Struct2.Name, in.Struct2.Name)
}

type Struct4 struct {
	Name2 string
}

type Struct3 struct {
	Struct4
	Name1 string
}

func TestEmbedding(t *testing.T) {
	in := &Struct3{
		Name1: "hello",
		Struct4: Struct4{
			Name2: "world",
		},
	}
	var out Struct3

	b, err := msgpack.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	err = msgpack.Unmarshal(b, &out)
	if err != nil {
		t.Fatal(err)
	}
	if out.Name1 != in.Name1 {
		t.Fatalf("")
	}
	if out.Name2 != in.Name2 {
		t.Fatalf("")
	}
}

func TestEmptyTimeMarshalWithInterface(t *testing.T) {
	a := time.Time{}
	b, err := msgpack.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var out interface{}
	err = msgpack.Unmarshal(b, &out)
	if err != nil {
		t.Fatal(err)
	}
	name, _ := out.(time.Time).Zone()
	if name != "UTC" {
		t.Fatal("Got wrong timezone")
	}

	var out2 time.Time
	err = msgpack.Unmarshal(b, &out2)
	if err != nil {
		t.Fatal(err)
	}
	name, _ = out2.Zone()
	if name != "UTC" {
		t.Fatal("Got wrong timezone")
	}
}

func (t *MsgpackTest) TestSliceNil() {
	in := [][]*int{nil}
	var out [][]*int

	t.Nil(t.enc.Encode(in))
	t.Nil(t.dec.Decode(&out))
	t.Equal(out, in)
}

//------------------------------------------------------------------------------

//------------------------------------------------------------------------------

func TestNoPanicOnUnsupportedKey(t *testing.T) {
	data := []byte{0x81, 0x81, 0xa1, 0x78, 0xc3, 0xc3}

	_, err := msgpack.NewDecoder(bytes.NewReader(data)).DecodeTypedMap()
	require.EqualError(t, err, "msgpack: unsupported map key: map[string]interface {}")
}

func TestMapDefault(t *testing.T) {
	in := map[string]interface{}{
		"foo": "bar",
		"hello": map[string]interface{}{
			"foo": "bar",
		},
	}
	b, err := msgpack.Marshal(in)
	require.Nil(t, err)

	var out map[string]interface{}
	err = msgpack.Unmarshal(b, &out)
	require.Nil(t, err)
	require.Equal(t, in, out)
}

func TestRawMessage(t *testing.T) {
	type In struct {
		Foo map[string]interface{}
	}

	type Out struct {
		Foo msgpack.RawMessage
	}

	type Out2 struct {
		Foo interface{}
	}

	b, err := msgpack.Marshal(&In{
		Foo: map[string]interface{}{
			"hello": "world",
		},
	})
	require.Nil(t, err)

	var out Out
	err = msgpack.Unmarshal(b, &out)
	require.Nil(t, err)

	var m map[string]string
	err = msgpack.Unmarshal(out.Foo, &m)
	require.Nil(t, err)
	require.Equal(t, map[string]string{
		"hello": "world",
	}, m)

	msg := new(msgpack.RawMessage)
	out2 := Out2{
		Foo: msg,
	}
	err = msgpack.Unmarshal(b, &out2)
	require.Nil(t, err)
	require.Equal(t, out.Foo, *msg)
}

func TestInterface(t *testing.T) {
	type Interface struct {
		Foo interface{}
	}

	in := Interface{Foo: "foo"}
	b, err := msgpack.Marshal(in)
	require.Nil(t, err)

	var str string
	out := Interface{Foo: &str}
	err = msgpack.Unmarshal(b, &out)
	require.Nil(t, err)
	require.Equal(t, "foo", str)
}

func TestNaN(t *testing.T) {
	in := float64(math.NaN())
	b, err := msgpack.Marshal(in)
	require.Nil(t, err)

	var out float64
	err = msgpack.Unmarshal(b, &out)
	require.Nil(t, err)
	require.True(t, math.IsNaN(out))
}

func TestSetSortMapKeys(t *testing.T) {
	in := map[string]interface{}{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
	}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.SetSortMapKeys(true)
	dec := msgpack.NewDecoder(&buf)

	err := enc.Encode(in)
	require.Nil(t, err)

	wanted := make([]byte, buf.Len())
	copy(wanted, buf.Bytes())
	buf.Reset()

	for i := 0; i < 100; i++ {
		err := enc.Encode(in)
		require.Nil(t, err)
		require.Equal(t, wanted, buf.Bytes())

		out, err := dec.DecodeMap()
		require.Nil(t, err)
		require.Equal(t, in, out)
	}
}

func TestSetOmitEmpty(t *testing.T) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.SetOmitEmpty(true)
	err := enc.Encode(EmbeddingPtrTest{})
	require.Nil(t, err)

	var t2 *EmbeddingPtrTest
	dec := msgpack.NewDecoder(&buf)
	err = dec.Decode(&t2)
	require.Nil(t, err)
	require.Nil(t, t2.Exported)

	type Nested struct {
		Foo string
		Bar string
	}
	type Item struct {
		X Nested
		Y *Nested
	}
	i := Item{}
	buf.Reset()
	err = enc.Encode(i)
	require.Nil(t, err)
	require.NotContains(t, buf.Bytes(), byte('X'))
	require.NotContains(t, buf.Bytes(), byte('Y'))

	i = Item{Y: &Nested{}}
	buf.Reset()
	err = enc.Encode(i)
	require.Nil(t, err)
	require.NotContains(t, buf.Bytes(), byte('X'))
	require.Contains(t, buf.Bytes(), byte('Y'))
}

type NullInt struct {
	Valid bool
	Int   int
}

func (i *NullInt) Set(j int) {
	i.Int = j
	i.Valid = true
}

func (i NullInt) IsZero() bool {
	return !i.Valid
}

func (i NullInt) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(i.Int)
}

func (i *NullInt) UnmarshalMsgpack(b []byte) error {
	if err := msgpack.Unmarshal(b, &i.Int); err != nil {
		return err
	}
	i.Valid = true
	return nil
}

type Secretive struct {
	Visible bool
	hidden  bool
}

type T struct {
	I NullInt `msgpack:",omitempty"`
	J NullInt
	// Secretive is not a "simple" struct because it has an hidden field.
	S Secretive `msgpack:",omitempty"`
}

func ExampleMarshal_ignore_simple_zero_structs_when_tagged_with_omitempty() {
	var t1 T
	raw, err := msgpack.Marshal(t1)
	if err != nil {
		panic(err)
	}
	var t2 T
	if err = msgpack.Unmarshal(raw, &t2); err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", t2)

	t2.I.Set(42)
	t2.S.hidden = true // won't be included because it is a hidden field
	raw, err = msgpack.Marshal(t2)
	if err != nil {
		panic(err)
	}
	var t3 T
	if err = msgpack.Unmarshal(raw, &t3); err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", t3)
	// Output: msgpack_test.T{I:msgpack_test.NullInt{Valid:false, Int:0}, J:msgpack_test.NullInt{Valid:true, Int:0}, S:msgpack_test.Secretive{Visible:false, hidden:false}}
	// msgpack_test.T{I:msgpack_test.NullInt{Valid:true, Int:42}, J:msgpack_test.NullInt{Valid:true, Int:0}, S:msgpack_test.Secretive{Visible:false, hidden:false}}
}

type (
	Value   interface{}
	Wrapper struct {
		Value Value `msgpack:"v,omitempty"`
	}
)

func TestEncodeWrappedValue(t *testing.T) {
	v := (*time.Time)(nil)
	c := &Wrapper{
		Value: v,
	}
	var buf bytes.Buffer
	require.Nil(t, msgpack.NewEncoder(&buf).Encode(v))
	require.Nil(t, msgpack.NewEncoder(&buf).Encode(c))
}

func TestPtrValueDecode(t *testing.T) {
	type Foo struct {
		Bar *int
	}

	b, err := msgpack.Marshal(Foo{})
	require.Nil(t, err)

	bar1 := 123
	foo := Foo{Bar: &bar1}

	err = msgpack.Unmarshal(b, &foo)
	require.Nil(t, err)
	require.Nil(t, foo.Bar)

	bar2 := 456
	b, err = msgpack.Marshal(Foo{Bar: &bar2})
	require.Nil(t, err)

	err = msgpack.Unmarshal(b, &foo)
	require.Nil(t, err)
	require.NotNil(t, foo.Bar)
	require.Equal(t, *foo.Bar, bar2)
}

// TestDecodeTypedMapSliceValuesAreIndependent is the regression test for #65.
// decodeTypedMapValue hoists the key/value reflect.Value slots out of the
// per-entry loop and zeroes them each iteration. Without the zeroing step,
// decodeSliceValue would reuse the previous iteration's backing array
// (since it reuses capacity when v.Cap() >= n), corrupting entries already
// written to the map.
func TestDecodeTypedMapSliceValuesAreIndependent(t *testing.T) {
	in := map[int][]int{
		1: {10, 11, 12},
		2: {20, 21, 22},
		3: {30, 31, 32},
	}

	b, err := msgpack.Marshal(in)
	require.Nil(t, err)

	var out map[int][]int
	require.Nil(t, msgpack.Unmarshal(b, &out))
	require.Equal(t, in, out)

	// Mutating one entry must not touch any other. If the decode loop
	// reused a single backing array, the entries would alias and this
	// assertion would fail.
	for i := range out[1] {
		out[1][i] = -1
	}
	require.Equal(t, []int{20, 21, 22}, out[2])
	require.Equal(t, []int{30, 31, 32}, out[3])
}

// TestDecodeTypedMapStructValuesAreIndependent covers the struct-value
// case of the same hoist bug. The struct carries a slice field so the
// aliasing mechanism fires through the hoisted value slot: decodeStructValue
// writes every field present in the payload, but the slice field itself is
// decoded via decodeSliceValue, which reuses the existing backing array
// when capacity is sufficient. Without per-iteration zeroing of the outer
// struct's slot, iteration 2's slice field would overwrite iteration 1's.
func TestDecodeTypedMapStructValuesAreIndependent(t *testing.T) {
	type kv struct {
		A int
		B []string
	}
	in := map[int]kv{
		1: {1, []string{"one", "uno"}},
		2: {2, []string{"two", "dos"}},
		3: {3, []string{"three", "tres"}},
	}

	b, err := msgpack.Marshal(in)
	require.Nil(t, err)

	var out map[int]kv
	require.Nil(t, msgpack.Unmarshal(b, &out))
	require.Equal(t, in, out)

	for i := range out[1].B {
		out[1].B[i] = "MUTATED"
	}
	require.Equal(t, []string{"two", "dos"}, out[2].B)
	require.Equal(t, []string{"three", "tres"}, out[3].B)
}
