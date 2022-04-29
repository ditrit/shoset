package shoset

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetLogLevel(lv string) {
	switch strings.ToLower(lv) {
	default:
		fallthrough
	case "info":
		log.Level(zerolog.InfoLevel)
	case "trace":
		log.Level(zerolog.TraceLevel)
	case "debug":
		log.Level(zerolog.DebugLevel)
	case "warn", "warning":
		log.Level(zerolog.WarnLevel)
	case "error":
		log.Level(zerolog.ErrorLevel)
	}
}

func LogWithCaller() {
	log.Logger = log.With().Caller().Logger()

}

func InitPrettyLogger(colored bool) {
	format := "2006-01-02T15:04:05.999Z07:00" // RFC3339 w/ millisecond
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: format, NoColor: !colored}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	log.Logger = log.Output(output)
	LogWithCaller()
	// log.Info().Str("foo", "bar").Msg("Logger initialised!")
	log.Info().Msg("Logger initialised!")
}

func Log(msg string) {
	log.Info().Msg(msg)
}
