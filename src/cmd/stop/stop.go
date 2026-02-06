// Package stop
package stop

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use: "stop",
		// TODO: change
		Short: "Stop cache server",
		// TODO: add foreground option?
		Long: `Stop cache server in the background`,
		Run:  stop,
	}

	return ptr
}

func stop(cmd *cobra.Command, args []string) {
}
