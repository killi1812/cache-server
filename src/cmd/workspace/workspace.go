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

func getServices() *service.WorkspaceSrv {
	var srv *service.WorkspaceSrv

	app.Invoke(func(s *service.WorkspaceSrv) {
		srv = s
	})
	return srv
}

// cache-server workspace create <workspace name> <cache name>
func create(cmd *cobra.Command, args []string) error {
	zap.S().Debugf("Trying to create workspace ...")
	wsName := args[0]
	cacheName := args[1]

	zap.S().Debugf("Parsed args: %v %v", wsName, cacheName)

	srv := getServices()

	token, err := auth.GenerateToken()
	if err != nil {
		zap.S().Errorf("Failed to generate token ")
		zap.S().Debug(err)
		return err
	}

	tmp := service.WorkspaceCreateArgs{
		WorkspaceName:   wsName,
		BinaryCacheName: cacheName,
		Token:           token,
	}
	worskpace, err := srv.Create(tmp)
	if err != nil {
		zap.S().Errorf("Failed to create workspace, err: %v", err)
		return err
	}

	fmt.Printf("Workspace Created Successfully\n")
	fmt.Printf("Name:       %s\n", worskpace.Name)
	fmt.Printf("Cache Name: %s\n", worskpace.BinaryCache.Name)
	fmt.Printf("Token:      %s\n", worskpace.Token)

	return nil
}

// cache-server workspace remove <workspace name>
func remove(cmd *cobra.Command, args []string) error {
	zap.S().Debugf("trying to delete binary cache ...")
	name := args[0]
	zap.S().Debugf("Parsed args: %v", name)

	serv := getServices()

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
	zap.S().Debugf("trying to list binary caches ...")

	// TODO: add json output
	serv := getServices()
	workspaces, err := serv.ReadAll()
	if err != nil {
		zap.S().Errorf("Failed to create workspace list, err: %+v", err)
		return err
	}

	zap.S().Debugf("Retrived %d workspaces", len(workspaces))

	fmt.Printf("Found %d workspaces:\n", len(workspaces))
	for _, wp := range workspaces {
		fmt.Printf("\t%s\n", wp.Name)
	}

	return nil
}

// cache-server workspace info <workspace name>
func info(cmd *cobra.Command, args []string) error {
	zap.S().Debugf("trying to read info of workspace ...")
	name := args[0]
	zap.S().Debugf("Parsed args: %v", name)

	serv := getServices()

	wp, err := serv.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read workspace, err: %+v", err)
		return err
	}

	zap.S().Debugf("Retrived workspace %s", name)
	tmpb := strings.Builder{}
	tmpe := json.NewEncoder(&tmpb)
	tmpe.SetIndent("", "   ")
	tmpe.Encode(wp)
	zap.S().Debug(tmpb.String())

	fmt.Printf("Name:       %s\n", wp.Name)
	if wp.BinaryCache != nil {
		fmt.Printf("Cache Name: %s\n", wp.BinaryCache.Name)
	} else {
		fmt.Printf("Cache:      null\n")
	}
	fmt.Printf("Token:      %s\n", wp.Token)
	fmt.Printf("Agents Cnt: %d\n", len(wp.Agents))

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
