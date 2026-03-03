package msgpack

import (
	"math"
	"reflect"
	"sort"
	"sync"

	"github.com/Basekick-Labs/msgpack/v6/msgpcode"
)

var sortedKeysPool = sync.Pool{
	New: func() interface{} {
		s := make([]string, 0, 16)
		return &s
	},
}

func getSortedKeys(n int) *[]string {
	sp := sortedKeysPool.Get().(*[]string)
	if cap(*sp) < n {
		*sp = make([]string, 0, n)
	} else {
		*sp = (*sp)[:0]
	}
	return sp
}

func putSortedKeys(sp *[]string) {
	if cap(*sp) > 1024 {
		return // don't retain oversized slices
	}
	sortedKeysPool.Put(sp)
}

func encodeMapValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.EncodeMapLen(v.Len()); err != nil {
		return err
	}

	iter := v.MapRange()
	for iter.Next() {
		if err := e.EncodeValue(iter.Key()); err != nil {
			return err
		}
		if err := e.EncodeValue(iter.Value()); err != nil {
			return err
		}
	}

	return nil
}

func encodeMapStringBoolValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.EncodeMapLen(v.Len()); err != nil {
		return err
	}

	var m map[string]bool
	if v.Type() == mapStringBoolType {
		m = v.Interface().(map[string]bool)
	} else {
		m = v.Convert(mapStringBoolType).Interface().(map[string]bool)
	}
	if e.flags&sortMapKeysFlag != 0 {
		return e.encodeSortedMapStringBool(m)
	}

	for mk, mv := range m {
		if err := e.EncodeString(mk); err != nil {
			return err
		}
		if err := e.EncodeBool(mv); err != nil {
			return err
		}
	}

	return nil
}

func encodeMapStringStringValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.EncodeMapLen(v.Len()); err != nil {
		return err
	}

	var m map[string]string
	if v.Type() == mapStringStringType {
		m = v.Interface().(map[string]string)
	} else {
		m = v.Convert(mapStringStringType).Interface().(map[string]string)
	}
	if e.flags&sortMapKeysFlag != 0 {
		return e.encodeSortedMapStringString(m)
	}

	for mk, mv := range m {
		if err := e.EncodeString(mk); err != nil {
			return err
		}
		if err := e.EncodeString(mv); err != nil {
			return err
		}
	}

	return nil
}

func encodeMapStringInterfaceValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}
	var m map[string]interface{}
	if v.Type() == mapStringInterfaceType {
		m = v.Interface().(map[string]interface{})
	} else {
		m = v.Convert(mapStringInterfaceType).Interface().(map[string]interface{})
	}
	if e.flags&sortMapKeysFlag != 0 {
		return e.EncodeMapSorted(m)
	}
	return e.EncodeMap(m)
}

func (e *Encoder) encodeMapStringString(m map[string]string) error {
	if m == nil {
		return e.EncodeNil()
	}
	if err := e.EncodeMapLen(len(m)); err != nil {
		return err
	}
	if e.flags&sortMapKeysFlag != 0 {
		return e.encodeSortedMapStringString(m)
	}
	for mk, mv := range m {
		if err := e.EncodeString(mk); err != nil {
			return err
		}
		if err := e.EncodeString(mv); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) EncodeMap(m map[string]interface{}) error {
	if m == nil {
		return e.EncodeNil()
	}
	if err := e.EncodeMapLen(len(m)); err != nil {
		return err
	}
	for mk, mv := range m {
		if err := e.EncodeString(mk); err != nil {
			return err
		}
		if err := e.Encode(mv); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) EncodeMapSorted(m map[string]interface{}) error {
	if m == nil {
		return e.EncodeNil()
	}
	if err := e.EncodeMapLen(len(m)); err != nil {
		return err
	}

	sp := getSortedKeys(len(m))
	keys := *sp

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		if err := e.EncodeString(k); err != nil {
			*sp = keys
			putSortedKeys(sp)
			return err
		}
		if err := e.Encode(m[k]); err != nil {
			*sp = keys
			putSortedKeys(sp)
			return err
		}
	}

	*sp = keys
	putSortedKeys(sp)
	return nil
}

func (e *Encoder) encodeSortedMapStringBool(m map[string]bool) error {
	sp := getSortedKeys(len(m))
	keys := *sp

	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if err := e.EncodeString(k); err != nil {
			*sp = keys
			putSortedKeys(sp)
			return err
		}
		if err := e.EncodeBool(m[k]); err != nil {
			*sp = keys
			putSortedKeys(sp)
			return err
		}
	}

	*sp = keys
	putSortedKeys(sp)
	return nil
}

func (e *Encoder) encodeSortedMapStringString(m map[string]string) error {
	sp := getSortedKeys(len(m))
	keys := *sp

	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if err := e.EncodeString(k); err != nil {
			*sp = keys
			putSortedKeys(sp)
			return err
		}
		if err := e.EncodeString(m[k]); err != nil {
			*sp = keys
			putSortedKeys(sp)
			return err
		}
	}

	*sp = keys
	putSortedKeys(sp)
	return nil
}

func (e *Encoder) EncodeMapLen(l int) error {
	if l < 16 {
		return e.writeCode(msgpcode.FixedMapLow | byte(l))
	}
	if l <= math.MaxUint16 {
		return e.write2(msgpcode.Map16, uint16(l))
	}
	return e.write4(msgpcode.Map32, uint32(l))
}

func encodeStructValue(e *Encoder, strct reflect.Value) error {
	structFields := structs.Fields(strct.Type(), e.structTag)
	if e.flags&arrayEncodedStructsFlag != 0 || structFields.AsArray {
		return encodeStructValueAsArray(e, strct, structFields.List)
	}
	fields := structFields.OmitEmpty(e, strct)

	if err := e.EncodeMapLen(len(fields)); err != nil {
		putFilteredFields(structFields, fields)
		return err
	}

	for _, f := range fields {
		if err := e.EncodeString(f.name); err != nil {
			putFilteredFields(structFields, fields)
			return err
		}
		if err := f.EncodeValue(e, strct); err != nil {
			putFilteredFields(structFields, fields)
			return err
		}
	}

	putFilteredFields(structFields, fields)
	return nil
}

// putFilteredFields returns a pooled filtered field slice.
// It is a no-op when the slice is the original fs.List (not pooled).
func putFilteredFields(fs *fields, filtered []*field) {
	if len(filtered) > 0 && len(fs.List) > 0 && &filtered[0] == &fs.List[0] {
		return
	}
	for i := range filtered {
		filtered[i] = nil
	}
	if cap(filtered) <= 64 {
		s := filtered
		filteredFieldsPool.Put(&s)
	}
}

func encodeStructValueAsArray(e *Encoder, strct reflect.Value, fields []*field) error {
	if err := e.EncodeArrayLen(len(fields)); err != nil {
		return err
	}
	for _, f := range fields {
		if err := f.EncodeValue(e, strct); err != nil {
			return err
		}
	}
	return nil
}
