package zap

type Field struct{}

type Logger struct{}

func NewNop() *Logger {
	return &Logger{}
}

func (l *Logger) Debug(msg string, fields ...Field) {}

func (l *Logger) Info(msg string, fields ...Field) {}

func (l *Logger) Warn(msg string, fields ...Field) {}

func (l *Logger) Error(msg string, fields ...Field) {}

