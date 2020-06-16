package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

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

func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) //skip first 3 caller Caller ,trace ,defer

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)    //获取对应的函数
		file, line := fn.FileLine(pc)  //获取调用该函数的文件名和行号
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line)) //空行，缩进
	}
	return str.String()
}
