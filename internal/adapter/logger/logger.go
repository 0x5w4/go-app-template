package logger

type Logger interface {
	NewInstance() Logger
	Field(key string, value any) Logger
	WithFields(fields map[string]any) Logger
	Logger() Logger
	Debug() LogEvent
	Info() LogEvent
	Warn() LogEvent
	Error() LogEvent
	Fatal() LogEvent
}

type LogEvent interface {
	Err(err error) LogEvent
	Field(key string, value any) LogEvent
	Msg(a ...any)
	Msgf(format string, a ...any)
}
