package main

import (
	"cephtower/backend/internal/configinit"
	"flag"
	"fmt"
	"os"
)

func main() {
	template := flag.String("template", "", "template path")
	target := flag.String("target", "", "target path")
	serverDir := flag.String("server-dir", "./app", "server directory")
	flag.Parse()
	if err := configinit.Create(*template, *target, *serverDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
