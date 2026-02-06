// Package agent
package agent

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:   "agent",
		Short: "Manage deployment agents",
		Run:   agent,
	}

	// TODO: add subcommands
	return ptr
}

func agent(cmd *cobra.Command, args []string) {
}

/*
		add                 Create agent
    remove              Remove agent
    list                List agents
    info                Display info about agent
*/
