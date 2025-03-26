package common

import (
	"io"
	"log"
	"os"
)

// Call after flag.Parse()
func InitLogging() {
	if !InfoEnabled {
		log.SetOutput(io.Discard)
	} else {
		log.SetOutput(os.Stdout)
	}

	if DebugEnabled {
		log.Println("🐞 Debug mode is ON")
	}
}

func Debugf(format string, v ...any) {
	if DebugEnabled {
		log.Printf("🐞 "+format, v...)
	}
}

func Infof(format string, v ...any) {
	if InfoEnabled {
		log.Printf("ℹ️ "+format, v...)
	}
}
