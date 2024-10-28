package logger

const LevelDebug = 5
const LevelExtra = 4
const LevelNormal = 3
const LevelWarn = 2
const LevelError = 1

var level int
var defaultLogger = New("")

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
	defaultLogger.Error(message, a...)
}

func Error2(err error) {
	defaultLogger.Error2(err)
}

func Warn(message string, a ...any) {
	defaultLogger.Warn(message, a...)
}

func Log(message string, a ...any) {
	defaultLogger.Log(message, a...)
}

func Info(message string, a ...any) {
	defaultLogger.Info(message, a...)
}

func Debug(message string, a ...any) {
	defaultLogger.Debug(message, a...)
}
