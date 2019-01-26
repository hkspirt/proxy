//----------------
//Func  :
//Author: xjh
//Date  : 2019/
//Note  :
//----------------
package util

import (
	"fmt"
	"log"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}

func format(tag string, v ...interface{}) {
	if len(v) > 1 {
		log.Output(3, fmt.Sprintf(fmt.Sprintf("%s %v", tag, v[0]), v[1:]...))
	} else {
		log.Output(3, fmt.Sprintf("%s %v", tag, v[0]))
	}
}

func LogDebug(v ...interface{}) {
	format("[D]", v...)
}

func LogInfo(v ...interface{}) {
	format("[I]", v...)
}

func LogWarn(v ...interface{}) {
	format("[W]", v...)
}

func LogError(v ...interface{}) {
	format("[E]", v...)
}

func LogFatal(v ...interface{}) {
	format("[F]", v...)
}
