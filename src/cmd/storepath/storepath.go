// Package storepath
package storepath

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:   "store-path",
		Short: "Manage deployment paths",
		Run:   storepath,
	}

	return ptr
}

func storepath(cmd *cobra.Command, args []string) {
}

/*
		list              List store paths
    delete            Delete store path
    info              Display info about store path
*/
