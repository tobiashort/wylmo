package must

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Must2[T any](val T, err error) T {
	Must(err)
	return val
}

func Must3[T1 any, T2 any](val1 T1, val2 T2, err error) (T1, T2) {
	Must(err)
	return val1, val2
}
