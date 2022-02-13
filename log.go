package feta

import "log"

func Log(m interface{}) {
	if Flags.Verbose {
		log.Println("DEBUG:", m)
	}
}

func Fatal(m interface{}) {
	log.Fatalln(m)
}
