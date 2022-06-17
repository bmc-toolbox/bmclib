package helper

import (
	"regexp"
	"runtime"
)

var basename = regexp.MustCompile(`^.+\.(.*$)`)

// WhosCalling returns the current caller of the functions
func WhosCalling() string {
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		return basename.ReplaceAllString(runtime.FuncForPC(pc).Name(), "${1}")
	}
	return "unknown"
}
