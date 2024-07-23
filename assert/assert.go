package assert

func NoErr(err error) {
	if err != nil {
		panic(err)
	}
}

func True(condition bool, message string) {
	if !condition {
		panic(message)
	}
}
