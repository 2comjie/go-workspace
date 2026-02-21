package convert

import (
	"fmt"
	"hutool/mathx"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

// String converts `any` to string.
// It's most commonly used converting function.
func String(any interface{}) string {
	if any == nil {
		return ""
	}
	switch value := any.(type) {
	case int:
		return strconv.Itoa(value)
	case int8:
		return strconv.Itoa(int(value))
	case int16:
		return strconv.Itoa(int(value))
	case int32:
		return strconv.Itoa(int(value))
	case int64:
		return strconv.FormatInt(value, 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case string:
		return value
	case []byte:
		return string(value)
	case time.Time:
		if value.IsZero() {
			return ""
		}
		return value.String()
	case *time.Time:
		if value == nil {
			return ""
		}
		return value.String()
	default:
		return fmt.Sprintf("%+v", value)
	}
}

// Int converts `any` to int.
func Int(any interface{}) int {
	if any == nil {
		return 0
	}
	if v, ok := any.(int); ok {
		return v
	}
	return int(Int64(any))
}

// Int8 converts `any` to int8.
func Int8(any interface{}) int8 {
	if any == nil {
		return 0
	}
	if v, ok := any.(int8); ok {
		return v
	}
	return int8(Int64(any))
}

// Int16 converts `any` to int16.
func Int16(any interface{}) int16 {
	if any == nil {
		return 0
	}
	if v, ok := any.(int16); ok {
		return v
	}
	return int16(Int64(any))
}

// Int32 converts `any` to int32.
func Int32(any interface{}) int32 {
	if any == nil {
		return 0
	}
	if v, ok := any.(int32); ok {
		return v
	}
	return int32(Int64(any))
}

// Int64 converts `any` to int64.
func Int64(any interface{}) int64 {
	if any == nil {
		return 0
	}
	switch value := any.(type) {
	case int:
		return int64(value)
	case int8:
		return int64(value)
	case int16:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	case uint:
		return int64(value)
	case uint8:
		return int64(value)
	case uint16:
		return int64(value)
	case uint32:
		return int64(value)
	case uint64:
		return int64(value)
	case float32:
		return int64(value)
	case float64:
		return int64(value)
	case bool:
		if value {
			return 1
		}
		return 0
	case string:
		v, err := strconv.Atoi(value)
		if err != nil {
			return 0
		}
		return int64(v)
	default:
		panic("not supported")
	}
}

// Uint converts `any` to uint.
func Uint(any interface{}) uint {
	if any == nil {
		return 0
	}
	if v, ok := any.(uint); ok {
		return v
	}
	return uint(Uint64(any))
}

// Uint8 converts `any` to uint8.
func Uint8(any interface{}) uint8 {
	if any == nil {
		return 0
	}
	if v, ok := any.(uint8); ok {
		return v
	}
	return uint8(Uint64(any))
}

// Uint16 converts `any` to uint16.
func Uint16(any interface{}) uint16 {
	if any == nil {
		return 0
	}
	if v, ok := any.(uint16); ok {
		return v
	}
	return uint16(Uint64(any))
}

// Uint32 converts `any` to uint32.
func Uint32(any interface{}) uint32 {
	if any == nil {
		return 0
	}
	if v, ok := any.(uint32); ok {
		return v
	}
	return uint32(Uint64(any))
}

// Uint64 converts `any` to uint64.
func Uint64(any interface{}) uint64 {
	return uint64(Int64(any))
}

func StringTobytes(s string) []byte {
	str := (*reflect.StringHeader)(unsafe.Pointer(&s))
	by := reflect.SliceHeader{
		Data: str.Data,
		Len:  str.Len,
		Cap:  str.Len,
	}
	//在把by从sliceheader转为[]byte类型
	return *(*[]byte)(unsafe.Pointer(&by))
}
func ByteToString(b []byte) string {
	by := (*reflect.SliceHeader)(unsafe.Pointer(&b))

	str := reflect.StringHeader{
		Data: by.Data,
		Len:  by.Len,
	}
	return *(*string)(unsafe.Pointer(&str))
}

func Bool(any interface{}) bool {
	bV, ok := any.(bool)
	if ok {
		return bV
	}
	sV, ok := any.(string)
	if ok {
		return mathx.In(sV, "True", "true", "TRUE")
	}
	return false
}
