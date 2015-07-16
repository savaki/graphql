package parse

import (
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	Debug = isDebugEnabled()
)

func isDebugEnabled() bool {
	v, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		v = false
	}

	if v {
		log.Println("debugging messages enagbled")
	}

	return v
}

func debug(format string, args ...interface{}) {
	if Debug {
		if strings.HasSuffix(format, "\n") {
			format = format + "\n"
		}
		log.Printf(format, args...)
	}
}
