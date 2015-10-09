package main

import (
	"torrentd/config"
	"os"
	"os/signal"
	"runtime/pprof"
)

func profile(cpuprofile string) {
	f, err := os.Create(cpuprofile)
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c

		config.L.Printf("profile: caught interrupt, stopping profiles")
		pprof.StopCPUProfile()

		os.Exit(0)
	}()
}