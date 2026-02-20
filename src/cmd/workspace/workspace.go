// Package workspace
package workspace

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serv *service.WorkspaceSrv

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
	zap.S().Infof("Trying to create workspace ...")
	wsName := args[0]
	cacheName := args[1]

	zap.S().Debugf("Parsed args: %v %v", wsName, cacheName)

	token, err := auth.GenerateToken()
	if err != nil {
		zap.S().Errorf("Failed to generate token, err: %v ", err)
		return err
	}

	tmp := service.WorkspaceCreateArgs{
		WorkspaceName:   wsName,
		BinaryCacheName: cacheName,
		Token:           token,
	}
	worskpace, err := serv.Create(tmp)
	if err != nil {
		zap.S().Errorf("Failed to create workspace, err: %v", err)
		return err
	}

	fmt.Printf("Workspace Created Successfully!\n")
	fmt.Printf("Name:       %s\n", worskpace.Name)
	fmt.Printf("Cache:      %s\n", worskpace.BinaryCache.Name)
	fmt.Printf("Token:      %s\n", worskpace.Token)

	return nil
}

// cache-server workspace remove <workspace name>
func remove(cmd *cobra.Command, args []string) error {
	zap.S().Infof("Trying to delete workspace ...")
	name := args[0]
	zap.S().Debugf("Parsed args: %v", name)

	if err := serv.Delete(name); err != nil {
		zap.S().Errorf("Failed to create cache token, err: %+v", err)
		return err
	}

	// Output for the user
	fmt.Printf("Workspace Removed Successfully!\n")
	return nil
}

// cache-server workspace list
func list(cmd *cobra.Command, args []string) error {
	zap.S().Infof("Trying to list workspaces ...")

	workspaces, err := serv.ReadAll()
	if err != nil {
		zap.S().Errorf("Failed to create workspace list, err: %+v", err)
		return err
	}

	zap.S().Debugf("Retrieved %d workspaces", len(workspaces))

	fmt.Printf("Found %d workspaces:\n", len(workspaces))
	for _, wp := range workspaces {
		fmt.Printf("\t%s\n", wp.Name)
	}

	return nil
}

// cache-server workspace info <workspace name>
func info(cmd *cobra.Command, args []string) error {
	zap.S().Infof("Trying to read info of workspace ...")
	name := args[0]
	zap.S().Debugf("Parsed args: %v", name)

	workspace, err := serv.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read workspace, err: %+v", err)
		return err
	}

	zap.S().Debugf("Retrieved workspace %s", name)
	tmpb := strings.Builder{}
	tmpe := json.NewEncoder(&tmpb)
	tmpe.SetIndent("", "   ")
	tmpe.Encode(workspace)
	zap.S().Debug(tmpb.String())

	fmt.Printf("Name:       %s\n", workspace.Name)
	if workspace.BinaryCache != nil {
		fmt.Printf("Cache Name: %s\n", workspace.BinaryCache.Name)
	} else {
		fmt.Printf("Cache:      null\n")
	}
	fmt.Printf("Token:      %s\n", workspace.Token)
	fmt.Printf("Agents Cnt: %d\n", len(workspace.Agents))

	return nil
}

// cache-server workspace cache <workspace name> <cache name>
func changeCache(cmd *cobra.Command, args []string) error {
	zap.S().Infof("Trying to update workspace cache ...")
	wsName := args[0]
	cacheName := args[1]
	zap.S().Debugf("Parsed args %v %v", wsName, cacheName)

	workspace, err := serv.UpdateCache(wsName, cacheName)
	if err != nil {
		zap.S().Errorf("Failed to update workspace %s cache to %s, err: %v ", wsName, cacheName, err)
		return err
	}

	fmt.Printf("Updated Workspace Cache Successfully!\n")
	fmt.Printf("Name:       %s\n", workspace.Name)
	fmt.Printf("Cache Name: %s\n", workspace.BinaryCache.Name)
	fmt.Printf("Token:      %s\n", workspace.Token)
	fmt.Printf("Agents Cnt: %d\n", len(workspace.Agents))

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

	zap.S().Debug("Running workspace setup ...")
	app.Invoke(func(s *service.WorkspaceSrv) {
		serv = s
	})

	return nil
}
