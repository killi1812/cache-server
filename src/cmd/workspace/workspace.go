// Package workspace
package workspace

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

/*
NewCmd creates a new workspace Command

	create              Create workspace
	delete              Delete workspace
	list                List workspaces
	info                Display info about workspace
	cache               Change workspace binary cache
*/
func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:               "workspace",
		Short:             "Manage deployment workspaces",
		Long:              "This command allows users to workspaces. It includes subcommands: create, delete, list, info, and cache.",
		PersistentPreRunE: setup,
		Run:               workspace,
	}

	ptr.AddCommand(
		&cobra.Command{
			Use:   "create [workspace name] [cache name]",
			Short: "Create a deployment workspace",
			Args:  cobra.ExactArgs(2),
			RunE:  create,
		},
		&cobra.Command{
			Use:   "delete [workspace name]",
			Short: "Delete a workspace and its agents",
			Args:  cobra.ExactArgs(1),
			RunE:  remove,
		},
		&cobra.Command{
			Use:   "list",
			Short: "List all workspaces",
			Args:  cobra.NoArgs,
			RunE:  list,
		},
		&cobra.Command{
			Use:   "info [workspace name]",
			Short: "Display detailed info about a workspace",
			Args:  cobra.ExactArgs(1),
			RunE:  info,
		},
		&cobra.Command{
			Use:   "cache [workspace name] [cache name]",
			Short: "Change the binary cache for a workspace",
			Args:  cobra.ExactArgs(2),
			RunE:  changeCache,
		},
	)

	return ptr
}

// cache-server workspace create <workspace name> <cache name>
func create(cmd *cobra.Command, args []string) error {
	wsName := args[0]
	cacheName := args[1]

	zap.S().Debugf("Creating workspace '%s' with cache '%s'", wsName, cacheName)
	// TODO: Save to DB and generate deployment token
	token := "generated-deploy-token-abc"
	zap.S().Debugf("Workspace created. Deployment Token: %s", token)
	return nil
}

// cache-server workspace remove <workspace name>
func remove(cmd *cobra.Command, args []string) error {
	wsName := args[0]
	zap.S().Debugf("Deleting workspace '%s' and all associated agents", wsName)
	// TODO: Cascading delete in DB
	return nil
}

// cache-server workspace list
func list(cmd *cobra.Command, args []string) error {
	zap.S().Debug("Listing all workspaces")
	// TODO: DB Query
	return nil
}

// cache-server workspace info <workspace name>
func info(cmd *cobra.Command, args []string) error {
	wsName := args[0]
	zap.S().Debugf("Fetching info for workspace: %s", wsName)
	// TODO: Return ID, Token, Name, and Cache Name
	return nil
}

// cache-server workspace cache <workspace name> <cache name>
func changeCache(cmd *cobra.Command, args []string) error {
	wsName := args[0]
	cacheName := args[1]
	zap.S().Debugf("Updating workspace '%s' to use cache '%s'", wsName, cacheName)
	// TODO: Update record in DB
	return nil
}

func workspace(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func setup(cmd *cobra.Command, args []string) error {
	// Attempt to run parent's setup (e.g., root command)

	parent := cmd.Parent().Parent()
	if parent != nil && parent.PersistentPreRun != nil {
		zap.S().Debugf("Running parent setup %d ...", parent.Use)
		parent.PersistentPreRun(parent, args)
	}

	zap.S().Debug("Running workspace setup...")
	return nil
}
