package main

// go fmt ./... && go vet ./... && go test && go run main/main.go -passPhrase XYZZY

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/erikbryant/shipahoy"
	_ "github.com/go-sql-driver/mysql"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "Enable profiling and write cpu profile to file")
	passPhrase = flag.String("passPhrase", "", "Passphrase to unlock API key(s)")
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

	if *passPhrase == "" {
		fmt.Println("You must specify -passPhrase")
		os.Exit(1)
	}

	err := shipahoy.Start(*passPhrase)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer shipahoy.Stop()

	// Only run the profile for a short while.
	if *cpuprofile != "" {
		time.Sleep(3 * 60 * time.Second)
		os.Exit(0)
	}

	// Let the scanners run forever.
	for {
		time.Sleep(24 * 60 * 60 * time.Second)
	}
}
