package mikado

import "log"

type logger func(format string, v ...interface{})

func defaultLogger(format string, v ...interface{}) {
	log.Printf("MIKADO: "+format, v...)
}
