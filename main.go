package main

import (
	"github.com/jgfranco17/lazyfile/cli/entrypoint"

	_ "embed" // Required for the //go:embed directive
)

//go:embed specs.json
var embeddedConfig []byte

func main() {
	entrypoint.Run(embeddedConfig)
}
