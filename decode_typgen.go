package msgpack

import (
	"reflect"
	"sync"
)

var cachedPools sync.Map // map[reflect.Type]*sync.Pool

func cachedValue(t reflect.Type) reflect.Value {
	v, ok := cachedPools.Load(t)
	if !ok {
		v, _ = cachedPools.LoadOrStore(t, &sync.Pool{
			New: func() interface{} {
				return reflect.New(t)
			},
		})
	}
	return v.(*sync.Pool).Get().(reflect.Value)
}

func (d *Decoder) newValue(t reflect.Type) reflect.Value {
	if d.flags&usePreallocateValues == 0 {
		return reflect.New(t)
	}

	return cachedValue(t)
}
