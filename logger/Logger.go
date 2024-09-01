package logger

import (
	"fmt"
	"os"
)

const LevelDebug = 5
const LevelExtra = 4
const LevelNormal = 3
const LevelWarn = 2
const LevelError = 1

var level int

func init() {
	level = LevelNormal
}

func SetLevel(newLevel int) {
	level = newLevel
}

func IncreaseLevel(increment int) {
	level += increment
}

func Error(message string, a ...any) {
	fmt.Fprintln(os.Stderr, formatMessage(LevelError, message, a...))
}

func Error2(err error) {
	fmt.Fprintln(os.Stderr, formatMessage(LevelError, err.Error()))
}

func Warn(message string, a ...any) {
	if level < LevelWarn {
		return
	}
	fmt.Fprintln(os.Stderr, formatMessage(LevelWarn, message, a...))
}

func Log(message string, a ...any) {
	if level < LevelNormal {
		return
	}
	fmt.Println(formatMessage(LevelNormal, message, a...))
}

func Info(message string, a ...any) {
	if level < LevelExtra {
		return
	}
	fmt.Println(formatMessage(LevelError, message, a...))
}

func Debug(message string, a ...any) {
	if level < LevelDebug {
		return
	}
	fmt.Println(formatMessage(LevelDebug, message, a...))
}

func formatMessage(msgLevel int, message string, a ...any) string {
	parsed := fmt.Sprintf(message, a...)

	if msgLevel == LevelNormal && level == LevelNormal {
		return parsed
	}

	var lvlStr string
	switch msgLevel {
	case LevelDebug:
		lvlStr = "DBG"
	case LevelError:
		lvlStr = "ERR"
	case LevelWarn:
		lvlStr = "WRN"
	case LevelExtra:
		lvlStr = "INF"
	case LevelNormal:
		lvlStr = "LOG"
	}

	return fmt.Sprintf("[%v] %v", lvlStr, parsed)
}
