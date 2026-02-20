// Package agent
package agent

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

// Package level service
var serv *service.AgentSrv

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
		// NOTE: for backwards compatibility
		&cobra.Command{
			Use:   "remove [agent name]",
			Short: "Remove an agent, left for backwards compatibility",
			Args:  cobra.ExactArgs(1),
			RunE:  remove,
		},
		&cobra.Command{
			Use:   "delete [agent name]",
			Short: "Remove an agent, same as remove",
			Args:  cobra.ExactArgs(1),
			RunE:  remove,
		},
	)

	return ptr
}

// cache-server agent add <agent name> <workspace name>
func add(cmd *cobra.Command, args []string) error {
	zap.S().Infof("Trying to add agent to workspace ...")
	agentName := args[0]
	wsName := args[1]
	zap.S().Debugf("Args: %+v", args)

	token, err := auth.GenerateToken()
	if err != nil {
		zap.S().Errorf("Failed to generate token, err: %v ", err)
		return err
	}

	tmp := service.AgentCreateArgs{
		AgentName:     agentName,
		WorkspaceName: wsName,
		Token:         token,
	}

	agent, err := serv.Create(tmp)
	if err != nil {
		zap.S().Errorf("Failed to create agent, err: %v", err)
		return err
	}

	fmt.Printf("Agent Created Successfully!\n")
	fmt.Printf("Name:       %s\n", agent.Name)
	fmt.Printf("Workspace:  %s\n", agent.Workspace.Name)
	fmt.Printf("Token:      %s\n", agent.Token)
	return nil
}

// cache-server agent list <workspace name>
func list(cmd *cobra.Command, args []string) error {
	zap.S().Infof("Trying to list agents ...")
	workspace := args[0]
	zap.S().Debugf("Parsed args %+v", args)

	agents, err := serv.ReadAll(workspace)
	if err != nil {
		zap.S().Errorf("Failed to read agents for workspace '%s', err: %v", workspace, err)
		return err
	}

	fmt.Printf("Agents for workspace '%s':\n", workspace)
	for _, agent := range agents {
		fmt.Printf("\t%s\n", agent.Name)
	}

	return nil
}

func info(cmd *cobra.Command, args []string) error {
	zap.S().Infof("Trying to read info for agent ...")
	name := args[0]
	zap.S().Debugf("Args: %+v", args)

	agent, err := serv.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read workspace, err: %+v", err)
		return err
	}

	zap.S().Debugf("Retrived workspace %s", name)
	tmpb := strings.Builder{}
	tmpe := json.NewEncoder(&tmpb)
	tmpe.SetIndent("", "   ")
	tmpe.Encode(agent)
	zap.S().Debug(tmpb.String())

	fmt.Printf("Name:       %s\n", agent.Name)
	if agent.Workspace != nil {
		fmt.Printf("Workspace:  %s\n", agent.Workspace.Name)
	} else {
		fmt.Printf("Workspace:null\n")
	}
	fmt.Printf("Token:      %s\n", agent.Token)

	return nil
}

// cache-server agent remove <agent name>
func remove(cmd *cobra.Command, args []string) error {
	zap.S().Infof("Trying to delete agent ...")
	name := args[0]
	zap.S().Debugf("Args: %+v", args)

	err := serv.Delete(name)
	if err != nil {
		zap.S().Errorf("Failed to remove agent '%s', err: %v", name, err)
		return err
	}

	fmt.Printf("Agent Removed Successfully!\n")
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
	app.Invoke(func(s *service.AgentSrv) {
		serv = s
	})

	return nil
}
