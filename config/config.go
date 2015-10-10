package config

import (
	"encoding/json"
	"os"
	"time"
	"log"
	"fmt"
	"strings"
)

type Config struct {
	Listen string
	Basedir string
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

func check() error {
	if len(C.Listen) == 0 {
		return fmt.Errorf("No listen-port given.")
	}
	if _, e := os.Stat(C.Basedir); os.IsNotExist(e) {
		return fmt.Errorf("Basedir does not exist: %s", C.Basedir)
	}
	if !strings.HasSuffix("/", C.Basedir) {
		C.Basedir += "/"
	}
	return nil
}

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
	return check()
}