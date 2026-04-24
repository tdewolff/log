package log

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Level int

const (
	NoneLevel Level = iota
	FatalLevel
	ErrorLevel
	WarningLevel
	AuditLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

type Target int

const (
	Terminal Target = iota
	File
	Journal
)

type Logger struct {
	w  io.Writer
	mu sync.Mutex

	Level    Level
	Target   Target
	Callback func(Level, string)
}

func (log *Logger) Write(level Level, msg string, calldepth int) {
	if log.Level < level {
		return
	} else if strings.HasSuffix(msg, "\n") {
		msg = msg[:len(msg)-1]
	}

	_, file, line, ok := runtime.Caller(calldepth + 1)
	if !ok {
		file = "???"
		line = 0
	} else if slash := strings.LastIndexByte(file, '/'); slash != -1 {
		file = file[slash+1:]
	}
	msg = fmt.Sprintf("%s:%d: %s", file, line, msg)
	if log.Callback != nil {
		log.Callback(level, msg)
	}

	switch log.Target {
	case Terminal:
		now := time.Now().Format("15:04:05")
		log.mu.Lock()
		defer log.mu.Unlock()
		switch level {
		case FatalLevel:
			fmt.Fprintf(log.w, "\033[41m\033[30m%s FATAL:\033[0m %s\n", now, msg)
		case ErrorLevel:
			fmt.Fprintf(log.w, "\033[31m%s ERROR:\033[0m %s\n", now, msg)
		case WarningLevel:
			fmt.Fprintf(log.w, "\033[33m%s WARN :\033[0m %s\n", now, msg)
		case AuditLevel:
			fmt.Fprintf(log.w, "\033[34m%s AUDIT:\033[0m %s\n", now, msg)
		case InfoLevel:
			fmt.Fprintf(log.w, "%s INFO : %s\n", now, msg)
		case DebugLevel:
			fmt.Fprintf(log.w, "\033[37m%s DEBUG:\033[0m %s\n", now, msg)
		case TraceLevel:
			fmt.Fprintf(log.w, "\033[90m%s TRACE:\033[0m %s\n", now, msg)
		}
	case File:
		now := time.Now().Format("2006-01-02 15:04:05")
		log.mu.Lock()
		defer log.mu.Unlock()
		switch level {
		case FatalLevel:
			fmt.Fprintf(log.w, "%s FATAL: %s\n", now, msg)
		case ErrorLevel:
			fmt.Fprintf(log.w, "%s ERROR: %s\n", now, msg)
		case WarningLevel:
			fmt.Fprintf(log.w, "%s WARN : %s\n", now, msg)
		case AuditLevel:
			fmt.Fprintf(log.w, "%s AUDIT: %s\n", now, msg)
		case InfoLevel:
			fmt.Fprintf(log.w, "%s INFO : %s\n", now, msg)
		case DebugLevel:
			fmt.Fprintf(log.w, "%s DEBUG: %s\n", now, msg)
		case TraceLevel:
			fmt.Fprintf(log.w, "%s TRACE: %s\n", now, msg)
		}
	case Journal:
		log.mu.Lock()
		defer log.mu.Unlock()
		switch level {
		case FatalLevel:
			fmt.Fprintf(log.w, "<2>FATAL: %s\n", msg)
		case ErrorLevel:
			fmt.Fprintf(log.w, "<3>ERROR: %s\n", msg)
		case WarningLevel:
			fmt.Fprintf(log.w, "<4>WARN : %s\n", msg)
		case AuditLevel:
			fmt.Fprintf(log.w, "<6>AUDIT: %s\n", msg)
		case InfoLevel:
			fmt.Fprintf(log.w, "<6>INFO : %s\n", msg)
		case DebugLevel:
			fmt.Fprintf(log.w, "<7>DEBUG: %s\n", msg)
		case TraceLevel:
			fmt.Fprintf(log.w, "<7>TRACE: %s\n", msg)
		}
	}
}

var Log = Logger{os.Stderr, sync.Mutex{}, TraceLevel, Terminal, nil}
var AuditLog = Logger{os.Stderr, sync.Mutex{}, AuditLevel, Terminal, nil}

func fdIsJournalStream(fd int) bool {
	journalStream := os.Getenv("JOURNAL_STREAM")
	if journalStream == "" {
		return false
	}
	var expectedStat syscall.Stat_t
	if _, err := fmt.Sscanf(journalStream, "%d:%d", &expectedStat.Dev, &expectedStat.Ino); err != nil {
		return false
	}
	var stat syscall.Stat_t
	if err := syscall.Fstat(fd, &stat); err != nil {
		return false
	}
	return stat.Dev == expectedStat.Dev && stat.Ino == expectedStat.Ino
}

func init() {
	if fdIsJournalStream(syscall.Stderr) {
		Log.Target = Journal
		AuditLog.Target = Journal
	} else if stat, err := os.Stderr.Stat(); err == nil && stat.Mode()&os.ModeCharDevice == 0 {
		// output piped to program or file
		Log.Target = File
		AuditLog.Target = File
	}
}

type Loggers struct {
	logFile, auditFile *os.File
}

func Config(level, logFilename, auditFilename string) *Loggers {
	switch strings.ToLower(level) {
	case "none":
		Log.Level = NoneLevel
		AuditLog.Level = NoneLevel
	case "fatal", "critical", "emergency":
		Log.Level = FatalLevel
		AuditLog.Level = NoneLevel
	case "error", "alert":
		Log.Level = ErrorLevel
		AuditLog.Level = NoneLevel
	case "warn", "warning":
		Log.Level = WarningLevel
		AuditLog.Level = NoneLevel
	default:
		Log.Level = WarningLevel
	case "info", "information", "notice":
		Log.Level = InfoLevel
	case "debug":
		Log.Level = DebugLevel
	case "trace":
		// no-op
	}
	l := &Loggers{}
	if logFilename != "" {
		if f, err := os.Create(logFilename); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: create logfile: %v", err)
			os.Exit(1)
		} else {
			Log.w = f
			Log.Target = File
			l.logFile = f
		}
	}
	if auditFilename != "" {
		if f, err := os.Create(auditFilename); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: create audit logfile: %v", err)
			os.Exit(1)
		} else {
			AuditLog.w = f
			AuditLog.Target = File
			l.auditFile = f
		}
	}
	return l
}

func (l *Loggers) Close() {
	if l.auditFile != nil {
		if err := l.auditFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: close audit logfile: %v", err)
		}
	}
	if l.logFile != nil {
		if err := l.logFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: close logfile: %v", err)
		}
	}
}

func Fatal(v ...any) {
	Log.Write(FatalLevel, fmt.Sprintln(v...), 1)
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	Log.Write(FatalLevel, fmt.Sprintf(format, v...), 1)
	os.Exit(1)
}

func Error(v ...any) {
	Log.Write(ErrorLevel, fmt.Sprintln(v...), 1)
}

func Errorf(format string, v ...any) {
	Log.Write(ErrorLevel, fmt.Sprintf(format, v...), 1)
}

func Warning(v ...any) {
	Log.Write(WarningLevel, fmt.Sprintln(v...), 1)
}

func Warningf(format string, v ...any) {
	Log.Write(WarningLevel, fmt.Sprintf(format, v...), 1)
}

func Audit(v ...any) {
	AuditLog.Write(AuditLevel, fmt.Sprintln(v...), 1)
}

func Auditf(format string, v ...any) {
	AuditLog.Write(AuditLevel, fmt.Sprintf(format, v...), 1)
}

func Info(v ...any) {
	Log.Write(InfoLevel, fmt.Sprintln(v...), 1)
}

func Infof(format string, v ...any) {
	Log.Write(InfoLevel, fmt.Sprintf(format, v...), 1)
}

func Debug(v ...any) {
	Log.Write(DebugLevel, fmt.Sprintln(v...), 1)
}

func Debugf(format string, v ...any) {
	Log.Write(DebugLevel, fmt.Sprintf(format, v...), 1)
}

func Trace(v ...any) {
	Log.Write(TraceLevel, fmt.Sprintln(v...), 1)
}

func Tracef(format string, v ...any) {
	Log.Write(TraceLevel, fmt.Sprintf(format, v...), 1)
}

func NewLogLogger(level Level) *log.Logger {
	return log.New(&logWriter{level}, "", 0)
}

type logWriter struct {
	level Level
}

func (l *logWriter) Write(b []byte) (int, error) {
	if l.level == AuditLevel {
		AuditLog.Write(l.level, string(b), 3)
	} else {
		Log.Write(l.level, string(b), 3)
	}
	return len(b), nil
}

func NewSlogLogger() *slog.Logger {
	return slog.New(&slogHandler{})
}

type slogHandler struct {
	groups string
	attrs  string
}

func (l *slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (l *slogHandler) Handle(ctx context.Context, rec slog.Record) error {
	if rec.Level < slog.LevelInfo {
		Log.Write(DebugLevel, rec.Message+l.attrs, 3)
	} else if rec.Level < slog.LevelWarn {
		Log.Write(InfoLevel, rec.Message+l.attrs, 3)
	} else if rec.Level < slog.LevelError {
		Log.Write(WarningLevel, rec.Message+l.attrs, 3)
	} else {
		Log.Write(ErrorLevel, rec.Message+l.attrs, 3)
	}
	return nil
}

func (l *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	l2 := &slogHandler{
		groups: l.groups,
		attrs:  l.attrs,
	}
	for _, attr := range attrs {
		l2.attrs += " " + attr.Key + "=" + fmt.Sprint(attr.Value)
	}
	return l2
}

func (l *slogHandler) WithGroup(group string) slog.Handler {
	l2 := &slogHandler{
		groups: l.groups + group + ".",
		attrs:  l.attrs,
	}
	return l2
}
