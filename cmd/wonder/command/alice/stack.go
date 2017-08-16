package alice

import (
	"fmt"
	"runtime"
)

func getStack() string {
	trace := make([]byte, 1024)
	runtime.Stack(trace, true)
	return fmt.Sprintf("%s", trace)
}
