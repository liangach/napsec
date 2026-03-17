package main

import (
	"fmt"
	"github.com/liangach/napsec/cmd/napsec/commands"
	"os"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
