// Package cache
package cache

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:   "cache",
		Short: "Manage caches",
		Run:   cache,
	}

	return ptr
}

func cache(cmd *cobra.Command, args []string) {
}

/*
		create              Create binary cache
    start               Start binary cache
    stop                Stop binary cache
    delete              Delete binary cache
    update              Update binary cache
    list                List binary caches
    info                Display info about binary cache
*/
