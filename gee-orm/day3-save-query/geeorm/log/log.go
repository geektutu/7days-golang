package log

import (
	"log"
	"os"
)

var (
	// Error is a logger for logging error messages
	Error = log.New(os.Stdout, "[error] ", log.LstdFlags|log.Lshortfile)
	// Info is a logger for logging normal messages
	Info = log.New(os.Stdout, "[info ] ", log.LstdFlags|log.Lshortfile)
)
