package main

import (
	"encoding/json"
	"flag"
	"fmt"
)

// Flags
var flagVerbose = flag.Bool("v", false, "Enable verbose output")

// SVR global server object
var SVR Server

// ROSTER global player list
var ROSTER Roster

func main() {
	flag.Parse()

	ROSTER = newRoster()

	SVR = newServer(false, 300)
	SVR.Start()

	netListen()
}

func prettyPrint(i interface{}) {
	b, _ := json.MarshalIndent(i, "", "  ")
	fmt.Printf("\n%s\n", b)
}
