package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jeffcail/go-load-plugins/loader"
)

func main() {
	l, err := loader.InitLoader()
	if err != nil {
		log.Fatal(err)
	}
	defer l.Destroy()

	for {
		for _, name := range l.Plugins() {
			if err := l.CompileAndRun(name); err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
			}
		}
	}
}
