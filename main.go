package main

import (
	"fmt"
	"os"

	"github.com/mpartipilo/autovendor/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "install":
		err = cmd.Install(os.Args[2:])
	case "uninstall":
		err = cmd.Uninstall(os.Args[2:])
	case "run":
		err = cmd.Run(os.Args[2:])
	case "version":
		fmt.Println("autovendor", cmd.Version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "autovendor: unknown command %q\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "autovendor: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`autovendor — automatic Go vendor sync after git operations

Usage:
  autovendor install [path]     Install git hooks into a Go repository
  autovendor uninstall [path]   Remove autovendor hooks from a repository
  autovendor run <hook> [args]  Run vendor sync (called by git hooks)
  autovendor version            Print version
  autovendor help               Show this help

Examples:
  autovendor install                 Install hooks in current directory
  autovendor install ~/src/myproject Install hooks in a specific repo
  autovendor uninstall               Remove hooks from current directory

Learn more: https://github.com/mpartipilo/autovendor
`)
}
