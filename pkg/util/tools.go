package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// countdown per second with ascii escape sequences, e.g. ("blastoff in ", 3, "...")
func Countdown(prefix string, n uint, suffix string) {
	for i := n; i > 0; i-- {
		fmt.Printf("%s%d%s", prefix, i, suffix)
		time.Sleep(time.Second * 1) // sleep for 1 second
		fmt.Print("\r\033[K")       // return to beginning of line and ascii escape to clear line
	}
}

func Dumps(obj any) string {
	data, _ := json.MarshalIndent(obj, "", "  ")
	return string(data)
}

func IsFile(path string) bool {
	if info, err := os.Stat(path); err != nil {
		return false
	} else {
		return !info.IsDir() // it should exist and NOT be a directory
	}
}

// defer this to make any function recoverable and log the error
func Recoverable(silent bool) {
	if r := recover(); r != nil {
		if !silent {
			var evt = log.Error()
			switch r := r.(type) {
			case error:
				evt = evt.AnErr("recovered", r)
			case string:
				evt = evt.AnErr("recovered", errors.New(r))
			default:
				evt = evt.Any("recovered", r)
			}
			evt.Stack().Msg("client panicked")
		}
	}
}
