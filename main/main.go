package main

import (
	"flag"
	"github.com/erikbryant/ship_ahoy"
	"log"
	"os"
	"runtime/pprof"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	passPhrase = flag.String("passPhrase", "", "Passphrase to unlock API key")
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	ship_ahoy.Doit(*passPhrase, *cpuprofile)
}
