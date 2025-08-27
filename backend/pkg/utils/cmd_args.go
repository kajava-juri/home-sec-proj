package utils

import (
	"flag"
)

type CommandArguments struct {
	CleanDb bool
}

var CmdArgs = CommandArguments{}

func ParseFlags() {
	testc := flag.Bool("clean", false, "Clean the database")
	flag.Parse()
	CmdArgs.CleanDb = *testc
}
