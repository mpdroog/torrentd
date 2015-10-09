package config

import (
	"encoding/json"
	"os"
	"time"
	"log"
)

type Config struct {
	Listen string
}

var (
	C           Config
	Verbose     bool
	Hostname    string
	ConfigPath  string
	Stopping    bool
	L *log.Logger

	// Metrics
	Appstart    time.Time
)

func Init(f string) error {
	Appstart = time.Now()
	ConfigPath = f
	r, e := os.Open(f)
	if e != nil {
		return e
	}
	if e := json.NewDecoder(r).Decode(&C); e != nil {
		return e
	}

	Hostname, e = os.Hostname()
	if e != nil {
		return e
	}

	L = log.New(os.Stdout, "", log.LstdFlags)
	return nil
}