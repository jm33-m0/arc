package arc

import (
	"log"
	"os"
)

var DEBUG = os.Getenv("ARC_DEBUG") == "true"

func logging(fmt_str string, a ...interface{}) {
	if DEBUG {
		log.Printf(fmt_str, a...)
	}
}
