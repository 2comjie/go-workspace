package safe

import "hutool/logx"

func Run(fn func()) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("safe run panic: %v", r)
			}
		}()
		fn()
	}()
}
