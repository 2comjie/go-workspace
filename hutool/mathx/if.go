package mathx

func If[T any](pass bool, trueValue T, falseValue T) T {
	if pass {
		return trueValue
	} else {
		return falseValue
	}
}
