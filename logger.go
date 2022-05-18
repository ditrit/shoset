package shoset

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// SetLogLevel switches between log levels
func SetLogLevel(lv string) {
	switch strings.ToLower(lv) {
	default:
		fallthrough
	case INFO:
		log.Logger = log.Level(zerolog.InfoLevel)
	case TRACE:
		log.Logger = log.Level(zerolog.TraceLevel)
	case DEBUG:
		log.Logger = log.Level(zerolog.DebugLevel)
	case WARN:
		log.Logger = log.Level(zerolog.WarnLevel)
	case ERROR:
		log.Logger = log.Level(zerolog.ErrorLevel)
	}
}

// LogWithCaller adds file and line number to log
func LogWithCaller() {
	log.Logger = log.With().Caller().Logger()
}

// InitPrettyLogger overrides log with prettier syntax
func InitPrettyLogger(colored bool) {
	// define TimeFormat with predefined layout RFC3339Nano
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339Nano, NoColor: !colored}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	log.Logger = log.Output(output)
	LogWithCaller()
}

// Log : default log
func Log(msg string) {
	log.Info().Msg(msg)
}
