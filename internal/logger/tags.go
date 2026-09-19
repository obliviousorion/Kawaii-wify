package logger

import (
	"fmt"
	"log"
)

const (
	TagBoot  = "[BOOT ]"
	TagState = "[STATE]"
	TagAuth  = "[AUTH ]"
	TagNet   = "[NET  ]"
	TagIPC   = "[IPC  ]"
	TagWarn  = "[WARN ]"
	TagError = "[ERROR]"
	TagFatal = "[FATAL]"
)

func Boot(format string, v ...any)  { log.Printf(TagBoot+" "+format, v...) }
func State(format string, v ...any) { log.Printf(TagState+" "+format, v...) }
func Auth(format string, v ...any)  { log.Printf(TagAuth+" "+format, v...) }
func Net(format string, v ...any)   { log.Printf(TagNet+" "+format, v...) }
func IPC(format string, v ...any)   { log.Printf(TagIPC+" "+format, v...) }
func Warn(format string, v ...any)  { log.Printf(TagWarn+" "+format, v...) }
func Error(format string, v ...any) { log.Printf(TagError+" "+format, v...) }
func Fatal(format string, v ...any) { log.Fatalf(TagFatal+" "+format, v...) }

// Transition logs state machine changes with standard syntax:
// [STATE] Transition: PreviousState -> NextState (reason)
func Transition(from, to fmt.Stringer, reason string) {
	if reason != "" {
		State("Transition: %s -> %s (%s)", from, to, reason)
	} else {
		State("Transition: %s -> %s", from, to)
	}
}
