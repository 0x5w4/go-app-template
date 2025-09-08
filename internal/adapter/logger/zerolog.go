package logger

import (
	"fmt"
	"maps"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type zerologLogger struct {
	debug  bool
	logger zerolog.Logger
	fields map[string]any
}

func NewZerologLogger(debug bool) Logger {
	output := zerolog.ConsoleWriter{
		Out: os.Stdout,
		FormatTimestamp: func(i any) string {
			t, ok := i.(string)
			if !ok {
				return "[???]"
			}
			parsed, err := time.Parse(time.RFC3339, t)
			if err != nil {
				return "[invalid time]"
			}
			return fmt.Sprintf("[%s]", parsed.Format("2006/01/02 15:04:05"))
		},
		FormatLevel: func(i any) string {
			return fmt.Sprintf("[%s]", strings.ToUpper(fmt.Sprintf("%v", i)))
		},
		FormatMessage: func(i any) string {
			return fmt.Sprintf("%v,", i)
		},
		FormatFieldName: func(i any) string {
			return fmt.Sprintf("%s:", i)
		},
		FormatFieldValue: func(i any) string {
			return fmt.Sprintf(" %v,", i)
		},
		FormatErrFieldName: func(i any) string {
			return fmt.Sprintf("%s:", i)
		},
		FormatErrFieldValue: func(i any) string {
			return fmt.Sprintf(" %v,", i)
		},
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			zerolog.MessageFieldName,
		},
		FieldsOrder: []string{
			"component",
		},
		TimeFormat: time.RFC3339,
	}
	baseLogger := zerolog.New(output).With().Timestamp().Logger()

	return &zerologLogger{
		logger: baseLogger,
		fields: make(map[string]any),
		debug:  debug,
	}
}

func (l *zerologLogger) NewInstance() Logger {
	return &zerologLogger{
		logger: l.logger,
		fields: copyFields(l.fields),
		debug:  l.debug,
	}
}

func (l *zerologLogger) Field(key string, value any) Logger {
	newLogger := l.NewInstance().(*zerologLogger)
	newLogger.fields[key] = value

	return newLogger
}

func (l *zerologLogger) WithFields(fields map[string]any) Logger {
	newLogger := l.NewInstance().(*zerologLogger)
	maps.Copy(newLogger.fields, fields)

	return newLogger
}

func (l *zerologLogger) Logger() Logger {
	return l
}

func (l *zerologLogger) Debug() LogEvent {
	return newZerologEvent(l.fields, l.logger.Debug())
}

func (l *zerologLogger) Info() LogEvent {
	return newZerologEvent(l.fields, l.logger.Info())
}

func (l *zerologLogger) Warn() LogEvent {
	return newZerologEvent(l.fields, l.logger.Warn())
}

func (l *zerologLogger) Error() LogEvent {
	return newZerologEvent(l.fields, l.logger.Error())
}

func (l *zerologLogger) Fatal() LogEvent {
	return newZerologEvent(l.fields, l.logger.Fatal())
}

type zerologEvent struct {
	event *zerolog.Event
}

func newZerologEvent(initialFields map[string]any, event *zerolog.Event) LogEvent {
	if len(initialFields) > 0 {
		event.Fields(initialFields)
	}

	return &zerologEvent{event: event}
}

func (e *zerologEvent) Err(err error) LogEvent {
	e.event.Err(err)
	return e
}

func (e *zerologEvent) Field(key string, value any) LogEvent {
	e.event.Any(key, value)
	return e
}

func (e *zerologEvent) Msg(v ...any) {
	e.event.Msg(fmt.Sprint(v...))
}

func (e *zerologEvent) Msgf(format string, v ...any) {
	e.event.Msgf(format, v...)
}

func copyFields(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	maps.Copy(dst, src)

	return dst
}
