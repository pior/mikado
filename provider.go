package mikado

import (
	"reflect"
	"runtime"
)

type provider struct {
	constructor interface{}
	name        string
	inputTypes  []reflect.Type
	outputTypes []reflect.Type
}

func newProvider(constructor interface{}) *provider {
	t := reflect.TypeOf(constructor)

	if t.Kind() != reflect.Func {
		panic("a constructor must be a function type")
	}

	p := &provider{
		constructor: constructor,
		name:        runtime.FuncForPC(reflect.ValueOf(constructor).Pointer()).Name(),
	}

	// outputTypes
	for i := 0; i < t.NumOut(); i++ {
		p.outputTypes = append(p.outputTypes, t.Out(i))
	}

	// inputTypes
	numIn := t.NumIn()
	if t.IsVariadic() && numIn > 0 {
		numIn-- // remove the variadic parameter
	}
	for i := 0; i < numIn; i++ {
		p.inputTypes = append(p.inputTypes, t.In(i))
	}

	return p
}

func (p *provider) call(inputs []reflect.Value) (outputs []reflect.Value) {
	return reflect.ValueOf(p.constructor).Call(inputs)
}

func _isErrorType(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*error)(nil)).Elem())
}

func findProvider(a *App, key typeKey) *provider {
	var p *provider
	for _, pp := range a.providers {
		for _, output := range pp.outputTypes {
			outputKey := typeKey{Type: output}
			if key == outputKey {
				p = pp
			}
		}
	}
	return p
}
