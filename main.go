package main

import (
	"fmt"
	"os"

	"flag"

	"github.com/xackery/aatracker/aa"
	"github.com/xackery/aatracker/dps"
	"github.com/xackery/aatracker/player"
	"github.com/xackery/aatracker/tracker"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	//all := flag.Bool("all", false, "Parse entire log file")
	new := flag.Bool("new", true, "Parse new log file")

	flag.Parse()
	if flag.NArg() < 1 {
		return fmt.Errorf("usage: %s <log file>, use -new to parse new data only, dps to enable dpsing", os.Args[0])
	}

	t, err := tracker.New(flag.Arg(0))
	if err != nil {
		return fmt.Errorf("tracker: %w", err)
	}

	err = player.New()
	if err != nil {
		return fmt.Errorf("player: %w", err)
	}

	_, err = aa.New()
	if err != nil {
		return fmt.Errorf("aa: %w", err)
	}

	_, err = dps.New()
	if err != nil {
		return fmt.Errorf("dps: %w", err)
	}

	if !*new {
		fmt.Println("Parsing entire log file")
	}

	err = t.Start(!*new)
	if err != nil {
		return fmt.Errorf("tracker start: %w", err)
	}

	select {}
}
