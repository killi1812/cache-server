// Package version provides basic version infromation for app
package version

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
	"github.com/killi1812/go-cache-server/app"
)

type VersionCmd struct{}

// Execute implements subcommands.Command.
func (v *VersionCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	fmt.Printf("Version: %s\n", app.Version)
	fmt.Printf("Build: %s\n", app.Build)
	fmt.Printf("Commit: %s\n", app.CommitHash)
	fmt.Printf("Build Time Stamp: %s\n", app.BuildTimestamp)
	return 0
}

// Name implements subcommands.Command.
func (v *VersionCmd) Name() string {
	return "version"
}

// SetFlags implements subcommands.Command.
func (v *VersionCmd) SetFlags(*flag.FlagSet) {
}

// Synopsis implements subcommands.Command.
func (v *VersionCmd) Synopsis() string {
	return "prints app version, build type, commit hash, and build time stamp"
}

// Usage implements subcommands.Command.
func (v *VersionCmd) Usage() string { return "version" }
