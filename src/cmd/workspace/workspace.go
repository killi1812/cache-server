// Package workspace
package workspace

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:   "workspace",
		Short: "Manage deployment workspace",
		Run:   workspace,
	}

	return ptr
}

func workspace(cmd *cobra.Command, args []string) {
}

/*
   create              Create workspace
   delete              Delete workspace
   list                List workspaces
   info                Display info about workspace
   cache               Change workspace binary cache
*/
