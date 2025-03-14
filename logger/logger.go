/*
   gobooru-downloader
   Copyright (C) 2025 Kasyanov Nikolay Alexeevich (Unbewohnte)

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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
