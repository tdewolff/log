package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	SeverityInfo    = "I"
	SeverityWarning = "W"
	SeverityError   = "E"
)

const (
	ColorGrey    = "\033[30;1m"
	ColorRed     = "\033[31;1m"
	ColorGreen   = "\033[32;1m"
	ColorYellow  = "\033[33;1m"
	ColorBlue    = "\033[34;1m"
	ColorMagenta = "\033[35;1m"
	ColorCyan    = "\033[36;1m"
	ColorWhite   = "\033[37;1m"
	ColorReset   = "\033[0m"
)

var DefaultLogger = Logger{os.Stderr, ""}

func Info(args ...interface{}) {
	DefaultLogger.print(SeverityInfo, fmt.Sprintln(args...))
}

func Infof(format string, args ...interface{}) {
	DefaultLogger.print(SeverityInfo, fmt.Sprintf(format, args...))
}

func Warning(args ...interface{}) {
	DefaultLogger.print(SeverityWarning, fmt.Sprintln(args...))
}

func Warningf(format string, args ...interface{}) {
	DefaultLogger.print(SeverityWarning, fmt.Sprintf(format, args...))
}

func Error(args ...interface{}) {
	DefaultLogger.print(SeverityError, fmt.Sprintln(args...))
}

func Errorf(format string, args ...interface{}) {
	DefaultLogger.print(SeverityError, fmt.Sprintf(format, args...))
}

func F(args ...interface{}) string {
	s := ""
	for i := 0; i+1 < len(args); i += 2 {
		if i != 0 {
			s += " "
		}
		s += fmt.Sprintf("%s%v%s=%s%v%s", ColorGrey, args[i], ColorReset, ColorYellow, args[i+1], ColorReset)
	}
	return s
}

type Logger struct {
	w      io.Writer
	prefix string
}

func New(prefix string) *Logger {
	if len(prefix) > 0 && prefix[len(prefix)-1] != ' ' {
		prefix += " "
	}
	return &Logger{
		os.Stderr,
		prefix,
	}
}

func (l *Logger) print(severity, msg string) {
	color := ""
	if severity == SeverityWarning || severity == SeverityError {
		color = ColorRed
		_, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "???"
			line = 1
		} else {
			slash := strings.LastIndex(file, "/")
			if slash >= 0 {
				file = file[slash+1:]
			}
		}
		fmt.Fprintf(l.w, "%s%s %s]%s %s(%d):%s %s%s", color, severity, time.Now().Format("01-02 15:04:05.000"), ColorCyan, file, line, ColorReset, l.prefix, msg)
	} else {
		fmt.Fprintf(l.w, "%s %s] %s%s", severity, time.Now().Format("01-02 15:04:05.000"), l.prefix, msg)
	}
}

func (l *Logger) Info(args ...interface{}) {
	l.print(SeverityInfo, fmt.Sprintln(args...))
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.print(SeverityInfo, fmt.Sprintf(format, args...))
}

func (l *Logger) Warning(args ...interface{}) {
	l.print(SeverityWarning, fmt.Sprintln(args...))
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	l.print(SeverityWarning, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(args ...interface{}) {
	l.print(SeverityError, fmt.Sprintln(args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.print(SeverityError, fmt.Sprintf(format, args...))
}
