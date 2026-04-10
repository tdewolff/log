# Log

- Automatically detect if log is directed to file or journald and adjust accordingly.
- Set colors for log levels in terminal.
- Add time and file + line number to each entry.

## Usage
```go
import "github.com/tdewolff/log"

loggers:=log.Config(log.WarningLevel, "logfile.txt", "auditfile.txt") // filenames can be empty to use terminal/journald
defer loggers.Close()

// exits with status 1
log.Fatal("example", data)
log.Fatalf("example %v", data)

log.Error("example", data)
log.Errorf("example %v", data)

log.Warning("example", data)
log.Warningf("example %v", data)

log.Info("example", data)
log.Infof("example %v", data)

log.Debug("example", data)
log.Debugf("example %v", data)

log.Trace("example", data)
log.Tracef("example %v", data)

log.Audit("example", data)
log.Auditf("example %v", data)

// if you need a loggers from the standard library to direct to this logger
logLogger := log.NewLogLogger(log.ErrorLevel)
slogLogger := log.NewSlogLogger()
```
