//go:build profile
// +build profile

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to file")

func startProfiling() func() {
	flag.Parse()

	var cpuFile *os.File

	if *cpuprofile != "" {
		var err error
		cpuFile, err = os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create CPU profile: %v\n", err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			fmt.Fprintf(os.Stderr, "could not start CPU profile: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "[profile] CPU profiling enabled, writing to %s\n", *cpuprofile)
	}

	return func() {
		if cpuFile != nil {
			pprof.StopCPUProfile()
			cpuFile.Close()
			fmt.Fprintf(os.Stderr, "[profile] CPU profile written to %s\n", *cpuprofile)
		}

		if *memprofile != "" {
			f, err := os.Create(*memprofile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not create memory profile: %v\n", err)
				return
			}
			defer f.Close()
			if err := pprof.WriteHeapProfile(f); err != nil {
				fmt.Fprintf(os.Stderr, "could not write memory profile: %v\n", err)
				return
			}
			fmt.Fprintf(os.Stderr, "[profile] Memory profile written to %s\n", *memprofile)
		}
	}
}
