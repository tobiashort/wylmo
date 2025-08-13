package assert

import (
	"fmt"
)

func Nil(v any, msg string) {
	if v != nil {
		panic(fmt.Errorf("%s", msg))
	}
}

func Nilf(v any, format string, args ...any) {
	if v != nil {
		panic(fmt.Errorf(format, args...))
	}
}

func NotNil(v any, msg string) {
	if v == nil {
		panic(fmt.Errorf("%s", msg))
	}
}

func NotNilf(v any, format string, args ...any) {
	if v == nil {
		panic(fmt.Errorf(format, args...))
	}
}

func True(cond bool, msg string) {
	if !cond {
		panic(fmt.Errorf("%s", msg))
	}
}

func Truef(cond bool, format string, args ...any) {
	if !cond {
		panic(fmt.Errorf(format, args...))
	}
}

func False(cond bool, msg string) {
	if cond {
		panic(fmt.Errorf("%s", msg))
	}
}

func Falsef(cond bool, format string, args ...any) {
	if cond {
		panic(fmt.Errorf(format, args...))
	}
}
