package reflectx

import "reflect"

func GenericTypeOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func TypeOf(v any) reflect.Type {
	return reflect.TypeOf(v)
}

func NewPointerIns(t reflect.Type) any {
	if t.Kind() != reflect.Ptr {
		panic("input reflect.Type must be a pointer")
	}
	targetType := t.Elem()
	newIns := reflect.New(targetType)
	return newIns.Interface()
}
