package mathx

func In[T comparable](v T, list ...T) bool {
	for _, l := range list {
		if v == l {
			return true
		}
	}
	return false
}
