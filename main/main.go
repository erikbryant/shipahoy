package main

import (
	"flag"
	"fmt"
	"github.com/erikbryant/shipahoy"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"runtime/pprof"
	"time"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "Enable profiling and write cpu profile to file")
	passPhrase = flag.String("passPhrase", "", "Passphrase to unlock API key")
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Println(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	shipahoy.Start(*passPhrase)
	defer shipahoy.Stop()

	if *cpuprofile != "" {
		time.Sleep(3 * 60 * time.Second)
		os.Exit(0)
	}

	// Let the scanners run forever.
	for {
		time.Sleep(24 * 60 * 60 * time.Second)
	}
}
