package server

import (
	"fmt"
	"runtime"
)

func trace() string {
	trace := make([]byte, 1024)
	runtime.Stack(trace, true)
	return fmt.Sprintf("%s", trace)
}
