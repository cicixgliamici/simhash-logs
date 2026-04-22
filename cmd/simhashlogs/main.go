package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	subcmd := os.Args[1]
	args := os.Args[2:]

	var code int
	switch subcmd {
	case "dedup":
		code = runDedup(args, os.Stdin, os.Stdout, os.Stderr)
	case "eval":
		code = runEval(args, os.Stdin, os.Stdout, os.Stderr)
	case "-h", "--help", "help":
		printUsage()
		code = 0
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", subcmd)
		printUsage()
		code = 2
	}
	os.Exit(code)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: simhashlogs <command> [options]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  dedup    Find near-duplicate logs (default behavior)\n")
	fmt.Fprintf(os.Stderr, "  eval     Evaluate LSH precision/recall vs Brute-force\n")
	fmt.Fprintf(os.Stderr, "\nRun 'simhashlogs <command> -h' for more details.\n")
}
