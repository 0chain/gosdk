package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	NONE  = 0
	FATAL = 1
	ERROR = 2
	INFO  = 3
	DEBUG = 4
)

const cBlack = "\u001b[30m"
const cRed = "\u001b[31m"
const cGreen = "\u001b[32m"
const cYellow = "\u001b[33m"
const cBlue = "\u001b[34m"
const cMagenta = "\u001b[35m"
const cCyan = "\u001b[36m"
const cWhite = "\u001b[37m"
const cReset = "\u001b[0m"

const (
	strFATAL = cRed + "[FATAL]  "
	strERROR = cRed + "[ERROR]  "
	strINFO  = "[INFO]   "
	strDEBUG = "[DEBUG]  "
)

type loggerIf interface {
	Init(lvl int)
	Debug(v ...interface{})
	Info(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
}

type Logger struct {
	lvl      int
	logDebug *log.Logger
	logInfo  *log.Logger
	logError *log.Logger
	logFatal *log.Logger
	fWriter  io.Writer
}

func (l *Logger) Init(lvl int) {
	l.SetLevel(lvl)
	l.logDebug = log.New(os.Stderr, "0chain-sdk: "+strDEBUG, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.logInfo = log.New(os.Stderr, "0chain-sdk: "+strINFO, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.logError = log.New(os.Stderr, "0chain-sdk: "+strERROR, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.logFatal = log.New(os.Stderr, "0chain-sdk: "+strFATAL, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func (l *Logger) SetLevel(lvl int) {
	l.lvl = lvl
}

func (l *Logger) SetLogFile(logFile io.Writer, verbose bool) {
	dLogs := []io.Writer{logFile}
	iLogs := []io.Writer{logFile}
	eLogs := []io.Writer{logFile}
	fLogs := []io.Writer{logFile}
	if verbose {
		dLogs = append(dLogs, os.Stderr)
		iLogs = append(iLogs, os.Stderr)
		eLogs = append(iLogs, os.Stderr)
		fLogs = append(iLogs, os.Stderr)
	}
	l.logDebug = log.New(io.MultiWriter(dLogs...), "0chain-sdk: "+strDEBUG, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.logInfo = log.New(io.MultiWriter(iLogs...), "0chain-sdk: "+strINFO, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.logError = log.New(io.MultiWriter(eLogs...), "0chain-sdk: "+strERROR, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.logFatal = log.New(io.MultiWriter(fLogs...), "0chain-sdk: "+strFATAL, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func (l *Logger) Debug(v ...interface{}) {
	if l.lvl >= DEBUG {
		l.logDebug.Output(2, fmt.Sprint(v))
	}
}

func (l *Logger) Info(v ...interface{}) {
	if l.lvl >= INFO {
		l.logInfo.Output(2, fmt.Sprint(v))
	}
}

func (l *Logger) Error(v ...interface{}) {
	if l.lvl >= ERROR {
		l.logError.Output(2, fmt.Sprint(v)+cReset)
	}
}

func (l *Logger) Fatal(v ...interface{}) {
	if l.lvl >= FATAL {
		l.logFatal.Output(2, fmt.Sprint(v)+cReset)
	}
}

func (l *Logger) Close() {
	if c, ok := l.fWriter.(io.Closer); ok && c != nil {
		c.Close()
	}
}
