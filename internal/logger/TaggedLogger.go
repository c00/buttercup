package logger

import (
	"fmt"
	"os"
)

func New(tag string) TaggedLogger {
	if tag != "" {
		tag = fmt.Sprintf("[%v] ", tag)
	}
	return TaggedLogger{
		tag: tag,
	}
}

type TaggedLogger struct {
	tag string
}

func (l TaggedLogger) Error(message string, a ...any) {
	fmt.Fprintln(os.Stderr, l.formatMessage(LevelError, message, a...))
}

func (l TaggedLogger) Error2(err error) {
	fmt.Fprintln(os.Stderr, l.formatMessage(LevelError, err.Error()))
}

func (l TaggedLogger) Warn(message string, a ...any) {
	if level < LevelWarn {
		return
	}
	fmt.Fprintln(os.Stderr, l.formatMessage(LevelWarn, message, a...))
}

func (l TaggedLogger) Log(message string, a ...any) {
	if level < LevelNormal {
		return
	}
	fmt.Println(l.formatMessage(LevelNormal, message, a...))
}

func (l TaggedLogger) Info(message string, a ...any) {
	if level < LevelExtra {
		return
	}
	fmt.Println(l.formatMessage(LevelError, message, a...))
}

func (l TaggedLogger) Debug(message string, a ...any) {
	if level < LevelDebug {
		return
	}
	fmt.Println(l.formatMessage(LevelDebug, message, a...))
}

func (l TaggedLogger) formatMessage(msgLevel int, message string, a ...any) string {
	parsed := fmt.Sprintf(message, a...)

	if msgLevel == LevelNormal && level == LevelNormal {
		return parsed
	}

	//Just return an empty line
	if parsed == "" {
		return parsed
	}

	var lvlStr string
	switch msgLevel {
	case LevelDebug:
		lvlStr = "DEBUG"
	case LevelError:
		lvlStr = "ERROR"
	case LevelWarn:
		lvlStr = "WARNING"
	case LevelExtra:
		lvlStr = "INFO"
	case LevelNormal:
		lvlStr = "LOG"
	}

	return fmt.Sprintf("%v%v: %v", l.tag, lvlStr, parsed)
}
