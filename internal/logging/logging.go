package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

func Setup(verbosity int) {
	level := zerolog.InfoLevel
	if verbosity >= 2 {
		level = zerolog.TraceLevel
	} else if verbosity == 1 {
		level = zerolog.DebugLevel
	}
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	logger = zerolog.New(writer).Level(level).With().Timestamp().Logger()
}

func L() *zerolog.Logger {
	return &logger
}
