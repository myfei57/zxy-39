package main

import (
	"flag"
	"time"
)

// Config holds the process options of the PipeWatch console.
type Config struct {
	Addr        string
	DataDir     string
	ScanInterval time.Duration
}

// LoadConfig parses the command line flags.
func LoadConfig() Config {
	var config Config
	flag.StringVar(&config.Addr, "addr", ":8080", "HTTP listen address")
	flag.StringVar(&config.DataDir, "data", "./data", "file persistence root")
	flag.DurationVar(&config.ScanInterval, "scan-interval", 10*time.Second, "scan cycle interval")
	flag.Parse()
	return config
}
