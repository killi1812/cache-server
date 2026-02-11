// Package agent
package agent

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

/*
NewCmd creates a new agent command

			add                 Create agent
	    remove              Remove agent
	    list                List agents
	    info                Display info about agent
*/
func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:               "agent",
		Short:             "Manage deployment agents",
		Long:              "This command provides management of deployment agents. It contains subcommands add, remove, list, and info.",
		PersistentPreRunE: setup,
		Run:               agent,
	}

	ptr.AddCommand(
		&cobra.Command{
			Use:   "add [agent name] [workspace name]",
			Short: "Create a new agent entry",
			Args:  cobra.MinimumNArgs(2),
			RunE:  add,
		},
		&cobra.Command{
			Use:   "list [workspace name]",
			Short: "List agents in a specific workspace",
			Args:  cobra.ExactArgs(1),
			RunE:  list,
		},
		&cobra.Command{
			Use:   "info [agent name]",
			Short: "Display detailed info about an agent",
			Args:  cobra.ExactArgs(1),
			RunE:  info,
		},
		&cobra.Command{
			Use:   "remove [agent name]",
			Short: "Remove an agent",
			Args:  cobra.ExactArgs(1),
			RunE:  remove,
		},
	)

	return ptr
}

// cache-server agent add <agent name> <workspace name>
func add(cmd *cobra.Command, args []string) error {
	agentName := args[0]
	workspace := args[1]

	zap.S().Debugf("Adding agent '%s' to workspace '%s'", agentName, workspace)
	// TODO: Generate auth token and save to DB
	token := "generated-auth-token-xyz"
	zap.S().Debugf("Agent created. Auth Token: %s", token)
	return nil
}

// cache-server agent list <workspace name>
func list(cmd *cobra.Command, args []string) error {
	workspace := args[0]
	zap.S().Debugf("Listing agents for workspace: %s", workspace)
	// TODO: DB Query
	return nil
}

func info(cmd *cobra.Command, args []string) error {
	agentName := args[0]
	zap.S().Debugf("Fetching info for agent: %s", agentName)
	// TODO: Show name, token, and workspace
	return nil
}

// cache-server agent remove <agent name>
func remove(cmd *cobra.Command, args []string) error {
	agentName := args[0]
	zap.S().Debugf("Removing agent: %s (Unique check implied)", agentName)
	// TODO: Delete from DB
	return nil
}

func agent(cmd *cobra.Command, args []string) {
	cmd.Help()
}

// setup for agent subcommands
func setup(cmd *cobra.Command, args []string) error {
	// run parent setup

	parent := cmd.Parent().Parent()
	if parent != nil && parent.PersistentPreRun != nil {
		zap.S().Debugf("Running parent setup %d ...", parent.Use)
		parent.PersistentPreRun(parent, args)
	}

	zap.S().Debugf("Running agent setup...")
	return nil
}
