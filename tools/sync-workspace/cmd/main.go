package main

import (
	"fmt"
	"os"

	"cephtower/tools/sync-workspace/internal/workspace"
)

func main() {
	if err := workspace.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "sync-workspace: %v\n", err)
		os.Exit(1)
	}
}
