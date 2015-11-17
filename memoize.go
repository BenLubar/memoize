// Package memoize caches return values of functions.
package memoize

import (
	"reflect"
	"sync"
)

var interfaceType = reflect.TypeOf(new(interface{})).Elem()
var valueType = reflect.TypeOf(new(call))

type call struct {
	wait     <-chan struct{}
	results  []reflect.Value
	panicked reflect.Value
}

// Memoize takes a function and returns a function of the same type. The
// returned function remembers the return value(s) of the function call.
// Any pointer values will be used as an address, so functions that modify
// their arguments or programs that modify returned values will not work.
//
// The returned function is safe to call from multiple goroutines if the
// original function is. Panics are handled, so calling panic from a function
// will call panic with the same value on future invocations with the same
// arguments.
func Memoize(fn interface{}) interface{} {
	v := reflect.ValueOf(fn)
	t := v.Type()

	keyType := reflect.ArrayOf(t.NumIn(), interfaceType)
	cache := reflect.MakeMap(reflect.MapOf(keyType, valueType))
	var mtx sync.Mutex

	return reflect.MakeFunc(t, func(args []reflect.Value) (results []reflect.Value) {
		key := reflect.New(keyType).Elem()
		for i, v := range args {
			vi := v.Interface()
			key.Index(i).Set(reflect.ValueOf(&vi).Elem())
		}
		mtx.Lock()
		val := cache.MapIndex(key)
		if val.IsValid() {
			mtx.Unlock()
			c := val.Interface().(*call)
			<-c.wait
			if c.panicked.IsValid() {
				panic(c.panicked.Interface())
			}
			return c.results
		}
		w := make(chan struct{})
		c := &call{wait: w}
		cache.SetMapIndex(key, reflect.ValueOf(c))
		mtx.Unlock()

		panicked := true
		defer func() {
			if panicked {
				p := recover()
				c.panicked = reflect.ValueOf(p)
				close(w)
				panic(p)
			}
		}()

		if t.IsVariadic() {
			results = v.CallSlice(args)
		} else {
			results = v.Call(args)
		}
		panicked = false
		c.results = results
		close(w)

		return
	}).Interface()
}
