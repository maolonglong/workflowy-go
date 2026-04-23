package main

import (
	"os"

	"github.com/maolonglong/workflowy-go/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
