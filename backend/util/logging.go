package util

import (
    "fmt"
    "io"
    "os"
    "runtime"
    "strings"
    "time"

    "github.com/labstack/echo/v4"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "gopkg.in/natefinch/lumberjack.v2"
)

// RequestLoggerMiddleware attaches a contextual zerolog.Logger to echo.Context.
func RequestLoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Derive a logger with basic per-request context
        reqLogger := log.Logger.With().
            Str("ip", c.RealIP()).
            Str("method", c.Request().Method).
            Str("path", c.Path()).
            Logger()

        // Store in Echo context for downstream use
        c.Set("logger", reqLogger)

        // Continue the chain
        return next(c)
    }
}

// Retrieve the contextual logger if it exists, or fall back to global
func GetRequestLogger(c echo.Context) zerolog.Logger {
    if l, ok := c.Get("logger").(zerolog.Logger); ok {
        return l
    }
    return log.Logger
}

// levelFilterWriter implements zerolog.LevelWriter
type levelFilterWriter struct {
    writer io.Writer
    level  zerolog.Level
}

type errorCallerHook struct{}

func (h errorCallerHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
    if level < zerolog.ErrorLevel {
        return
    }

    // Loop through the stack to find the first non-zerolog frame
    for i := 3; i < 15; i++ { // check a reasonable depth
        _, file, line, ok := runtime.Caller(i)
        if !ok {
            break
        }

        // Skip frames from zerolog and runtime
        if strings.Contains(file, "rs/zerolog") || strings.Contains(file, "runtime/") {
            continue
        }

        // Extract just the file name
        short := file
        if idx := strings.LastIndex(file, "/"); idx != -1 {
            short = file[idx+1:]
        }

        e.Str("caller", fmt.Sprintf("%s:%d", short, line))
        break
    }
}

func (lw levelFilterWriter) Write(p []byte) (n int, err error) {
    return lw.writer.Write(p)
}

func (lw levelFilterWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
    // Only write if log level >= threshold
    if level >= lw.level {
        return lw.writer.Write(p)
    }
    return len(p), nil // pretend we wrote it
}

func SetupLogging() {
    // --- File writer (rotating logs) ---
    rotatingFile := &lumberjack.Logger{
        Filename:   "logs/monopoly_backend_log",
        MaxSize:    10, // MB
        MaxBackups: 5,
        MaxAge:     28, // days
        Compress:   true,
    }

    // --- Console writer (pretty human-readable) ---
    consoleWriter := zerolog.ConsoleWriter{
        Out:        os.Stdout,
        TimeFormat: time.RFC3339,
        NoColor:    false,
    }

    // --- Wrap each writer with its own minimum level ---
    consoleLevelWriter := levelFilterWriter{writer: consoleWriter, level: zerolog.TraceLevel}
    fileLevelWriter := levelFilterWriter{writer: rotatingFile, level: zerolog.InfoLevel}

    // --- Multi-level writer ---
    multi := zerolog.MultiLevelWriter(consoleLevelWriter, fileLevelWriter)

    // --- Global logger ---
    log.Logger = zerolog.New(multi).
        Hook(errorCallerHook{}).
        With().
        Timestamp().
        Logger()

    zerolog.SetGlobalLevel(zerolog.TraceLevel) // global minimum; per-output filtering still applies
}
