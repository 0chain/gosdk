// Provides a simple logger for the SDK.
package logger

import (
	"fmt"
	"github.com/0chain/gosdk/core/version"
	"gopkg.in/natefinch/lumberjack.v2"
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

// Log global logger instance
var Log Logger

const cRed = "\u001b[31m"
const cReset = "\u001b[0m"

const (
	strFATAL = cRed + "[FATAL]  "
	strERROR = cRed + "[ERROR]  "
	strINFO  = "[INFO]   "
	strDEBUG = "[DEBUG]  "
)

type Logger struct {
	lvl      int
	prefix   string
	logDebug *log.Logger
	logInfo  *log.Logger
	logError *log.Logger
	logFatal *log.Logger
	fWriter  io.Writer
}

// Init - Initialize logging
func (l *Logger) Init(lvl int, prefix string) {
	l.SetLevel(lvl)
	l.prefix = prefix
	l.logDebug = log.New(os.Stderr, prefix+": "+strDEBUG, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.logInfo = log.New(os.Stderr, prefix+": "+strINFO, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.logError = log.New(os.Stderr, prefix+": "+strERROR, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.logFatal = log.New(os.Stderr, prefix+": "+strFATAL, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

// SetLevel - Configures the log level. Higher the number more verbose.
func (l *Logger) SetLevel(lvl int) {
	l.lvl = lvl
}

// syncPrefixes - syncs the logger prefixes
func syncPrefixes(maxPrefixLen int, loggers []*Logger) {
	for _, lgr := range loggers {
		if maxPrefixLen-len(lgr.prefix) > 0 {
			lgr.prefix = fmt.Sprintf("%-*s", maxPrefixLen, lgr.prefix)
		}
	}
}

// SyncLoggers syncs the loggers prefixes
//   - loggers is the list of loggers to sync
func SyncLoggers(loggers []*Logger) {
	maxPrefixLen := 0
	for _, lgr := range loggers {
		if len(lgr.prefix) > maxPrefixLen {
			maxPrefixLen = len(lgr.prefix)
		}
	}
	syncPrefixes(maxPrefixLen, loggers)
}

// SetLogFile - Writes log to the file. set verbose false disables log to os.Stderr
func (l *Logger) SetLogFile(logFile io.Writer, verbose bool) {
	dLogs := []io.Writer{logFile}
	iLogs := []io.Writer{logFile}
	eLogs := []io.Writer{logFile}
	fLogs := []io.Writer{logFile}
	if verbose {
		dLogs = append(dLogs, os.Stderr)
		iLogs = append(iLogs, os.Stderr)
		eLogs = append(eLogs, os.Stderr)
		fLogs = append(fLogs, os.Stderr)
	}
	l.logDebug = log.New(io.MultiWriter(dLogs...), l.prefix+" "+strDEBUG, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.logInfo = log.New(io.MultiWriter(iLogs...), l.prefix+" "+strINFO, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.logError = log.New(io.MultiWriter(eLogs...), l.prefix+" "+strERROR, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.logFatal = log.New(io.MultiWriter(fLogs...), l.prefix+" "+strFATAL, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func (l *Logger) Debug(v ...interface{}) {
	if l.lvl >= DEBUG {
		err := l.logDebug.Output(2, fmt.Sprint(v...))
		if err != nil {
			fmt.Printf("Error logging debug message: %v", err)
			return
		}
	}
}

func formatAnyValue(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case error:
		return v.Error()
	default:
		return fmt.Sprintf("%+v", v) // Default format for all types (including structs)
	}
}

func (l *Logger) Info(v ...interface{}) {
	if l.lvl >= INFO {
		// Format each argument to be readable
		for i, arg := range v {
			v[i] = formatAnyValue(arg)
		}

		err := l.logInfo.Output(2, fmt.Sprint(v...))
		if err != nil {
			fmt.Printf("Error logging info message: %v", err)
			return
		}
	}
}

func (l *Logger) Error(v ...interface{}) {
	if l.lvl >= ERROR {
		err := l.logError.Output(2, fmt.Sprint(v...)+cReset)
		if err != nil {
			fmt.Printf("Error logging error message: %v", err)
			return
		}
	}
}

func (l *Logger) Fatal(v ...interface{}) {
	if l.lvl >= FATAL {
		err := l.logFatal.Output(2, fmt.Sprint(v...)+cReset)
		if err != nil {
			fmt.Printf("Error logging fatal message: %v", err)
			return
		}
	}
}

func (l *Logger) Close() {
	if c, ok := l.fWriter.(io.Closer); ok && c != nil {
		err := c.Close()
		if err != nil {
			fmt.Printf("Error closing log file: %v", err)
			return
		}
	}
}

// Initialize common logger
var logging Logger

func GetLogger() *Logger {
	return &logging
}

func CloseLog() {
	logging.Close()
}

func init() {
	logging.Init(DEBUG, "0chain-core-sdk")
}

// SetLogLevel set the log level.
// lvl - 0 disabled; higher number (upto 4) more verbosity
func SetLogLevel(lvl int) {
	logging.SetLevel(lvl)
}

// SetLogFile - sets file path to write log
// verbose - true - console output; false - no console output
func SetLogFile(logFile string, verbose bool) {
	ioWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100, // MB
		MaxBackups: 5,   // number of backups
		MaxAge:     28,  //days
		LocalTime:  false,
		Compress:   false, // disabled by default
	}
	logging.SetLogFile(ioWriter, verbose)
	logging.Info("******* Wallet SDK Version:", version.VERSIONSTR, " ******* (SetLogFile)")
}
