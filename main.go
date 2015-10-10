package main

import (
	"os"
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
		fmt.Println(e.Error())
		os.Exit(1)
		return
	}
	if config.Verbose {
		fmt.Println(config.C)
	}

	Control()
}