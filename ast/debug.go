package ast

import (
	"encoding/json"
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

func (iter *iterator) dumpTokens() {
	log.Println("------------------------------------")
	for i := 0; i < len(iter.tokens); i++ {
		item := iter.peekN(Pos(i))
		log.Printf("%2d. %14s => %v\n", i, item.typ, item.val)
	}
}

func (iter *iterator) dumpOperations() {
	log.Println("------------------------------------")
	data, _ := json.MarshalIndent(iter.operations, "", "..")
	log.Println(string(data))
}
