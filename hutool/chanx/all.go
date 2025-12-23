package chanx

func DrainNow[T any](ch <-chan T) []T {
	var out []T
	for {
		select {
		case v := <-ch:
			out = append(out, v)
		default:
			return out // 没有立即可读的就返回
		}
	}
}
