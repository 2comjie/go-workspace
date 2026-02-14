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

func NewPointerIns2[T any]() T {
	t := GenericTypeOf[T]()
	ins := NewPointerIns(t)
	return ins.(T)
}

func NewIns[T any]() T {
	var zero T
	rType := reflect.TypeOf(zero)
	// 如果是指针类型，创建指向零值的新指针
	if rType.Kind() == reflect.Ptr {
		elemType := rType.Elem()
		return reflect.New(elemType).Interface().(T)
	}
	// 值类型直接返回零值
	return zero
}

func ConvertList[T any](list []any) []T {
	outList := make([]T, len(list))
	for _, v := range list {
		outList = append(outList, v.(T))
	}
	return outList
}
