package db_trace

import (
	"reflect"
)

type DBTrace struct{}

// fill fills up the nil callback with noOp callbacks
func (t *DBTrace) fill() {
	tv := reflect.ValueOf(t).Elem()
	noOpV := reflect.ValueOf(noOpTrace).Elem()
	structType := tv.Type()
	for i := 0; i < structType.NumField(); i++ {
		// get a field in this trace
		tf := tv.Field(i)
		// if this field is not a function, continue
		if tf.Type().Kind() != reflect.Func {
			continue
		}
		// if this field is nil
		if tf.IsNil() {
			// get the corresponding field from noOp
			noopf := noOpV.Field(i)
			if noopf.IsNil() {
				panic("NoOp Callback cannot be NIL")
			}
			// fill with the function from noOp
			tf.Set(noopf)
		}
	}
}

// compose modifies t such that it respects the previously-registered hooks in old,
// subject to the composition policy requested in t.Compose.
func (t *DBTrace) compose(old *DBTrace) {
	if old == nil {
		return
	}
	tv := reflect.ValueOf(t).Elem()
	ov := reflect.ValueOf(old).Elem()
	structType := tv.Type()
	for i := 0; i < structType.NumField(); i++ {
		tf := tv.Field(i)
		hookType := tf.Type()
		if hookType.Kind() != reflect.Func {
			continue
		}
		of := ov.Field(i)
		if of.IsNil() {
			continue
		}
		if tf.IsNil() {
			tf.Set(of)
			continue
		}

		// Make a copy of tf for tf to call. (Otherwise it
		// creates a recursive call cycle and stack overflows)
		tfCopy := reflect.ValueOf(tf.Interface())

		// We need to call both tf and of in some order.
		newFunc := reflect.MakeFunc(hookType, func(args []reflect.Value) []reflect.Value {
			tfCopy.Call(args)
			return of.Call(args)
		})
		tv.Field(i).Set(newFunc)
	}
}
