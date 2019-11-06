package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Log is the global logger for xp-bot.
// Should be used over printf when possible.
// See uber-go/zap for docs on how to use it.
var Log *zap.SugaredLogger

func init() {
	l, err := zap.NewProduction()
	if err != nil {
		fmt.Println("Failed to instantiate logger. CRITICAL ERROR")
		panic(-1)
	}
	Log = l.Sugar()
}
