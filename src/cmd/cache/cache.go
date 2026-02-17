// Package cache
package cache

import (
	"fmt"
	"strconv"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var retention int

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:               "cache",
		Short:             "Manage caches",
		PersistentPreRunE: setup,
		Run:               cache,
	}
	cr := &cobra.Command{
		Use:   "create [cache name] [port number]",
		Short: "Create a new binary cache",
		Args:  cobra.ExactArgs(2),
		RunE:  create,
	}
	cr.Flags().IntVarP(&retention, "retention", "r", 0, "Time to retain cache in days, 0 means forever")

	ptr.AddCommand(cr)

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

func create(cmd *cobra.Command, args []string) error {
	name := args[0]
	portstr := args[1]

	zap.S().Debugf("Parsed args: %v %v %v", name, portstr, retention)

	port, err := strconv.Atoi(portstr)
	if err != nil {
		return fmt.Errorf("port is not a number %s", portstr)
	}

	serv := getService()

	t, err := auth.GenerateToken()
	if err != nil {
		zap.S().Errorf("Failed to generate token ")
		zap.S().Debug(err)
		return nil
	}

	tmp := service.CreateCacheArgs{Name: name, Port: port, Retention: retention, Token: t}
	cache, err := serv.Create(tmp)
	if err != nil {
		zap.S().Errorf("Failed to create cache token %+v", err)
		return nil
	}

	zap.S().Debugf("Binary cache '%s' created successfully (ID: %d)", cache.Name, cache.ID)

	// Output for the user
	zap.S().Infof("Binary Cache Created Successfully!")
	zap.S().Infof("Name:      %s", cache.Name)
	zap.S().Infof("Port:      %d", cache.Port)
	zap.S().Infof("Token:     %s", cache.Token)
	// zap.S().Infof("Directory: %s", cachePath)
	if retention > 0 {
		zap.S().Infof("Retention: %d days", cache.Retention)
	}

	return nil
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

// getService gets the cache service
func getService() (s *service.CacheSrv) {
	app.Invoke(func(serv *service.CacheSrv) {
		s = serv
	})

	return
}
