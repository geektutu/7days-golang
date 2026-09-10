package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// print stack trace for debug
func trace(message string) string {
	// Allocate an array to store program counters
	var pcs [32]uintptr
	// Get call stack info, skip first 3 callers
	n := runtime.Callers(3, pcs[:])

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")

	// Create Frames object
	frames := runtime.CallersFrames(pcs[:n])
	for {
		// Get one frame per iteration
		frame, more := frames.Next()

		// Add file, line number and function name
		str.WriteString(fmt.Sprintf("\n\t%s:%d - %s",
			frame.File,
			frame.Line,
			frame.Function,
		))

		if !more {
			break
		}
	}
	return str.String()
}

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		c.Next()
	}
}
