package logger

import (
	"io"
	"log"
	"os"
)

// 3 basic loggers in global scope
var (
	// neutral information logger
	infoLog *log.Logger

	// warning-level information logger
	warningLog *log.Logger

	// errors information logger
	errorLog *log.Logger
)

func init() {
	infoLog = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	warningLog = log.New(os.Stdout, "[WARNING] ", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "[ERROR] ", log.Ldate|log.Ltime)
}

// Set up loggers to write to the given writer
func SetOutput(writer io.Writer) {
	if writer == nil {
		writer = io.Discard
	}
	infoLog.SetOutput(writer)
	warningLog.SetOutput(writer)
	errorLog.SetOutput(writer)
}

func Info(format string, a ...interface{}) {
	infoLog.Printf(format, a...)
}

func Warning(format string, a ...interface{}) {
	warningLog.Printf(format, a...)
}

func Error(format string, a ...interface{}) {
	errorLog.Printf(format, a...)
}
