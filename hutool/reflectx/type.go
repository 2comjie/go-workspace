package reflectx

import "reflect"

import "github.com/spf13/cast"

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

func IndirectValue(data any) reflect.Value {
	// Indirect 消除指针
	return reflect.Indirect(reflect.ValueOf(data))
}

func ConvertAndSet(v any, value reflect.Value) bool {
	realV := reflect.Indirect(value)
	rv := reflect.ValueOf(v)
	if rv.CanConvert(realV.Type()) {
		newV := rv.Convert(realV.Type())
		if realV.CanSet() {
			realV.Set(newV)
			return true
		}
	}

	ok, converted := ConvertToTarget(v, realV.Type())
	if ok {
		return ConvertAndSet(converted, value)
	}
	return false
}

func ConvertToTarget(v any, t reflect.Type) (bool, any) {
	switch t.Kind() {
	case reflect.Bool:
		return true, cast.ToBool(v)
	case reflect.Int:
		return true, cast.ToInt(v)
	case reflect.Int8:
		return true, cast.ToInt8(v)

	case reflect.Int16:
		return true, cast.ToInt16(v)

	case reflect.Int32:
		return true, cast.ToInt32(v)

	case reflect.Int64:
		return true, cast.ToInt64(v)

	case reflect.Uint:
		return true, cast.ToUint(v)

	case reflect.Uint8:
		return true, cast.ToUint8(v)

	case reflect.Uint16:
		return true, cast.ToUint16(v)

	case reflect.Uint32:
		return true, cast.ToUint32(v)

	case reflect.Uint64:
		return true, cast.ToUint64(v)
	case reflect.Float32:
		return true, cast.ToFloat32(v)

	case reflect.Float64:
		return true, cast.ToFloat64(v)
	case reflect.String:
		return true, v
	default:
		return false, v
	}
}
