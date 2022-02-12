package feta

import "log"

var verbose bool

func SetVerbose(v bool) {
	verbose = v
}

func Log(m interface{}) {
	if verbose {
		log.Println("DEBUG:", m)
	}
}

func Fatal(m interface{}) {
	log.Fatalln(m)
}
