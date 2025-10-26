package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/trains-io/z21.go"
)

func usage() {
	log.Printf("Usage: z21-client [-s server]\n")
	flag.PrintDefaults()
}

func showUsageAndExit(exitcode int) {
	usage()
	os.Exit(exitcode)
}

func main() {
	var url = flag.String("s", z21.DefaultURL, "The Z21 control station URL")
	var showHelp = flag.Bool("h", false, "Show help message")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if *showHelp {
		showUsageAndExit(0)
	}

	nc, err := z21.Connect(*url, z21.Verbose(true))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	m, err := nc.SendRcv(ctx, &z21.HwInfo{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("resp = %v\n", m)
}
