package msgpack

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/Basekick-Labs/msgpack/v6/msgpcode"
)

var errArrayStruct = errors.New("msgpack: number of fields in array-encoded struct has changed")

var (
	mapStringStringPtrType = reflect.TypeOf((*map[string]string)(nil))
	mapStringStringType    = mapStringStringPtrType.Elem()
	mapStringBoolPtrType   = reflect.TypeOf((*map[string]bool)(nil))
	mapStringBoolType      = mapStringBoolPtrType.Elem()
)

var (
	mapStringInterfacePtrType = reflect.TypeOf((*map[string]interface{})(nil))
	mapStringInterfaceType    = mapStringInterfacePtrType.Elem()
)

func decodeMapValue(d *Decoder, v reflect.Value) error {
	n, err := d.DecodeMapLen()
	if err != nil {
		return err
	}

	typ := v.Type()
	if n == -1 {
		v.Set(reflect.Zero(typ))
		return nil
	}

	if v.IsNil() {
		ln := n
		if d.flags&disableAllocLimitFlag == 0 {
			ln = min(ln, maxMapSize)
		}
		v.Set(reflect.MakeMapWithSize(typ, ln))
	}
	if n == 0 {
		return nil
	}

	return d.decodeTypedMapValue(v, n)
}

func (d *Decoder) decodeMapDefault() (interface{}, error) {
	if d.mapDecoder != nil {
		return d.mapDecoder(d)
	}

	n, err := d.DecodeMapLen()
	if err != nil {
		return nil, err
	}
	if n == -1 {
		return nil, nil
	}
	if n == 0 {
		return make(map[string]interface{}), nil
	}

	code, err := d.PeekCode()
	if err != nil {
		return nil, err
	}

	if msgpcode.IsString(code) {
		return d.decodeMapStringInterfaceN(n)
	}
	return d.decodeTypedMapN(n)
}

// DecodeMapLen decodes map length. Length is -1 when map is nil.
func (d *Decoder) DecodeMapLen() (int, error) {
	c, err := d.readCode()
	if err != nil {
		return 0, err
	}

	if msgpcode.IsExt(c) {
		if err = d.skipExtHeader(c); err != nil {
			return 0, err
		}

		c, err = d.readCode()
		if err != nil {
			return 0, err
		}
	}
	return d.mapLen(c)
}

func (d *Decoder) mapLen(c byte) (int, error) {
	if c == msgpcode.Nil {
		return -1, nil
	}
	if c >= msgpcode.FixedMapLow && c <= msgpcode.FixedMapHigh {
		return int(c & msgpcode.FixedMapMask), nil
	}
	if c == msgpcode.Map16 {
		size, err := d.uint16()
		return int(size), err
	}
	if c == msgpcode.Map32 {
		size, err := d.uint32()
		return int(size), err
	}
	return 0, unexpectedCodeError{code: c, hint: "map length"}
}

func decodeMapStringStringValue(d *Decoder, v reflect.Value) error {
	var mptr *map[string]string
	if v.Type() == mapStringStringType {
		mptr = v.Addr().Interface().(*map[string]string)
	} else {
		mptr = v.Addr().Convert(mapStringStringPtrType).Interface().(*map[string]string)
	}
	return d.decodeMapStringStringPtr(mptr)
}

func (d *Decoder) decodeMapStringStringPtr(ptr *map[string]string) error {
	size, err := d.DecodeMapLen()
	if err != nil {
		return err
	}
	if size == -1 {
		*ptr = nil
		return nil
	}

	m := *ptr
	if m == nil {
		ln := size
		if d.flags&disableAllocLimitFlag == 0 {
			ln = min(size, maxMapSize)
		}
		*ptr = make(map[string]string, ln)
		m = *ptr
	}

	for i := 0; i < size; i++ {
		mk, err := d.DecodeString()
		if err != nil {
			return err
		}
		mv, err := d.DecodeString()
		if err != nil {
			return err
		}
		m[mk] = mv
	}

	return nil
}

func decodeMapStringInterfaceValue(d *Decoder, v reflect.Value) error {
	var ptr *map[string]interface{}
	if v.Type() == mapStringInterfaceType {
		ptr = v.Addr().Interface().(*map[string]interface{})
	} else {
		ptr = v.Addr().Convert(mapStringInterfacePtrType).Interface().(*map[string]interface{})
	}
	return d.decodeMapStringInterfacePtr(ptr)
}

func (d *Decoder) decodeMapStringInterfacePtr(ptr *map[string]interface{}) error {
	m, err := d.DecodeMap()
	if err != nil {
		return err
	}
	*ptr = m
	return nil
}

func (d *Decoder) DecodeMap() (map[string]interface{}, error) {
	n, err := d.DecodeMapLen()
	if err != nil {
		return nil, err
	}
	if n == -1 {
		return nil, nil
	}
	return d.decodeMapStringInterfaceN(n)
}

func (d *Decoder) decodeMapStringInterfaceN(n int) (map[string]interface{}, error) {
	ln := n
	if d.flags&disableAllocLimitFlag == 0 && ln > maxMapSize {
		ln = maxMapSize
	}

	m := make(map[string]interface{}, ln)

	for i := 0; i < n; i++ {
		mk, err := d.DecodeString()
		if err != nil {
			return nil, err
		}
		mv, err := d.decodeInterfaceCond()
		if err != nil {
			return nil, err
		}
		m[mk] = mv
	}

	return m, nil
}

func (d *Decoder) DecodeUntypedMap() (map[interface{}]interface{}, error) {
	n, err := d.DecodeMapLen()
	if err != nil {
		return nil, err
	}

	if n == -1 {
		return nil, nil
	}

	m := make(map[interface{}]interface{}, n)

	for i := 0; i < n; i++ {
		mk, err := d.decodeInterfaceCond()
		if err != nil {
			return nil, err
		}

		mv, err := d.decodeInterfaceCond()
		if err != nil {
			return nil, err
		}

		m[mk] = mv
	}

	return m, nil
}

// DecodeTypedMap decodes a typed map. Typed map is a map that has a fixed type for keys and values.
// Key and value types may be different.
func (d *Decoder) DecodeTypedMap() (interface{}, error) {
	n, err := d.DecodeMapLen()
	if err != nil {
		return nil, err
	}
	if n <= 0 {
		return nil, nil
	}
	return d.decodeTypedMapN(n)
}

func (d *Decoder) decodeTypedMapN(n int) (interface{}, error) {
	key, err := d.decodeInterfaceCond()
	if err != nil {
		return nil, err
	}

	value, err := d.decodeInterfaceCond()
	if err != nil {
		return nil, err
	}

	keyType := reflect.TypeOf(key)

	if !keyType.Comparable() {
		return nil, fmt.Errorf("msgpack: unsupported map key: %s", keyType.String())
	}

	// Use interface{} as the value type so heterogeneous values (e.g.
	// nested maps with different inner types) decode without type errors.
	mapType := reflect.MapOf(keyType, interfaceType)

	ln := n
	if d.flags&disableAllocLimitFlag == 0 {
		ln = min(ln, maxMapSize)
	}

	mapValue := reflect.MakeMapWithSize(mapType, ln)
	mapValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))

	n--
	if err := d.decodeTypedMapValue(mapValue, n); err != nil {
		return nil, err
	}

	return mapValue.Interface(), nil
}

func (d *Decoder) decodeTypedMapValue(v reflect.Value, n int) error {
	if n == 0 {
		return nil
	}
	var (
		typ       = v.Type()
		keyType   = typ.Key()
		valueType = typ.Elem()
	)
	// Hoist the key and value slots out of the loop so we pay two
	// reflect.New allocations per map decode instead of 2N. SetMapIndex
	// copies the key and value, so the stored entries are unaffected by
	// subsequent mutation of mk/mv.
	//
	// Zeroing each iteration is required for correctness, not just
	// hygiene: value decoders like decodeSliceValue reuse an existing
	// slice's backing array when v.Cap() >= n, which would otherwise let
	// iteration 2 clobber iteration 1's already-stored slice.
	mk := d.newValue(keyType).Elem()
	mv := d.newValue(valueType).Elem()
	keyZero := reflect.Zero(keyType)
	valueZero := reflect.Zero(valueType)
	for i := 0; i < n; i++ {
		mk.Set(keyZero)
		if err := d.DecodeValue(mk); err != nil {
			return err
		}

		mv.Set(valueZero)
		if err := d.DecodeValue(mv); err != nil {
			return err
		}

		v.SetMapIndex(mk, mv)
	}

	return nil
}

func (d *Decoder) skipMap(c byte) error {
	n, err := d.mapLen(c)
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		if err := d.Skip(); err != nil {
			return err
		}
		if err := d.Skip(); err != nil {
			return err
		}
	}
	return nil
}

func decodeStructValue(d *Decoder, v reflect.Value) error {
	c, err := d.readCode()
	if err != nil {
		return err
	}

	n, err := d.mapLen(c)
	if err == nil {
		return d.decodeStruct(v, n)
	}

	var err2 error
	n, err2 = d.arrayLen(c)
	if err2 != nil {
		return err
	}

	if n <= 0 {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	fields := structs.Fields(v.Type(), d.structTag)
	if n != len(fields.List) {
		return errArrayStruct
	}

	for _, f := range fields.List {
		if err := f.DecodeValue(d, v); err != nil {
			return err
		}
	}

	return nil
}

func (d *Decoder) decodeStruct(v reflect.Value, n int) error {
	if n == -1 {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	fields := structs.Fields(v.Type(), d.structTag)
	for i := 0; i < n; i++ {
		name, err := d.decodeStringTemp()
		if err != nil {
			return err
		}

		if f := fields.Map[name]; f != nil {
			if err := f.DecodeValue(d, v); err != nil {
				return err
			}
			continue
		}

		if d.flags&disallowUnknownFieldsFlag != 0 {
			return fmt.Errorf("msgpack: unknown field %q", name)
		}
		if err := d.Skip(); err != nil {
			return err
		}
	}

	return nil
}
