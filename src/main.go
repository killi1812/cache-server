package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/cmd/version"
)

func init() {
	app.Setup()

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&version.VersionCmd{}, "")

	flag.Parse()
}

func main() {
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
