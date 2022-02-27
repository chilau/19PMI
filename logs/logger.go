package logs

import (
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"os"
	"sync"
	"time"
)

const (
	RFC3339Milli = "2006-01-02T15:04:05.999Z07:00"
)

var (
	logFileName string
)

type Logger struct {
	zerolog.Logger
	logFile *os.File
}

type Prefix struct {
	Title string
	Value string
}

func (l Logger) SetSource(source string) Logger {
	return l.AddPrefix(
		Prefix{
			Title: "src",
			Value: source,
		},
	)
}

func (l Logger) AddPrefix(prefixes ...Prefix) Logger {
	context := l.With()
	for _, prefix := range prefixes {
		context = context.Str(prefix.Title, prefix.Value)
	}

	l.UpdateContext(
		func(c zerolog.Context) zerolog.Context {
			return context
		},
	)

	return l
}

func (l Logger) TimeTrack(start time.Time, actionDescription string) {
	elapsed := time.Since(start)
	l.Info().Msgf(
		"%s took %s",
		actionDescription,
		elapsed,
	)
}

func (l Logger) LogConfig(configBytes []byte) {
	if l.logFile == nil {
		fmt.Println("appConfig:")
		fmt.Println(string(configBytes))
	} else {
		logString := fmt.Sprintf(
			"Application config:\n%s\n\n",
			string(configBytes),
		)

		_, _ = l.logFile.Write([]byte(logString))
	}
}

var doOnce sync.Once
var logger Logger

func SetLogFileName(fileName string) {
	logFileName = fileName
}

func GetLogger() *Logger {
	doOnce.Do(func() {
		zerolog.TimeFieldFormat = RFC3339Milli

		var writer io.Writer
		var logFile *os.File

		if logFileName != "" {
			var err error
			logFile, err = os.OpenFile(
				logFileName,
				os.O_RDWR|os.O_CREATE|os.O_APPEND,
				0666,
			)

			if err == nil {
				writer = logFile
			} else {
				fmt.Println(err.Error())
			}
		} else {
			writer = os.Stdout
		}

		if writer != nil {
			logger = Logger{
				Logger: zerolog.New(writer).
					With().
					Timestamp().
					Logger(),
				logFile: logFile,
			}
		} else {
			fmt.Println("writer is nil")
		}
	})

	return &logger
}
