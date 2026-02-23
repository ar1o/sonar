package main

import (
	"os"

	"github.com/ar1o/sonar/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
