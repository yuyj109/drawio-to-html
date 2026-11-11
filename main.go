package main

import (
	"fmt"
	"os"

	"github.com/yuyj109/drawio-to-html/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
