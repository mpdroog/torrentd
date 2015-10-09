package main

import (
	"fmt"
	"flag"
	"torrentd/config"
)

func main() {
	configPath := ""
	flag.BoolVar(&config.Verbose, "v", false, "Verbose-mode (log more)")
	flag.StringVar(&configPath, "c", "./config.json", "Path to config.json")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()

	if *cpuprofile != "" {
 		profile(*cpuprofile)
    }

	if e := config.Init(configPath); e != nil {
		panic(e)
	}
	if config.Verbose {
		fmt.Println(config.C)
	}
	if len(config.C.Listen) == 0 {
		panic("No listen-port given.")
	}
	Control()
}