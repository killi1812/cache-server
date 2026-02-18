// Package cache
package cache

import (
	"fmt"
	"strconv"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"github.com/killi1812/go-cache-server/util/objstor"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var retention int

/*
NewCmd creates a new cache command

			create              Create binary cache
	    start               Start binary cache
	    stop                Stop binary cache
	    delete              Delete binary cache
	    update              Update binary cache
	    list                List binary caches
	    info                Display info about binary cache
*/
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

	ptr.AddCommand(cr,
		&cobra.Command{
			Use:   "delete [cache name] ",
			Short: "delete a binary cache",
			Args:  cobra.ExactArgs(1),
			RunE:  remove,
		},
	)

	return ptr
}

func cache(cmd *cobra.Command, args []string) {}

func create(cmd *cobra.Command, args []string) error {
	name := args[0]
	portstr := args[1]

	zap.S().Debugf("Parsed args: %v %v %v", name, portstr, retention)

	port, err := strconv.Atoi(portstr)
	if err != nil {
		return fmt.Errorf("port is not a number %s", portstr)
	}

	serv, stor := getServices()

	t, err := auth.GenerateToken()
	if err != nil {
		zap.S().Errorf("Failed to generate token ")
		zap.S().Debug(err)
		return err
	}

	// TODO: create a space for binarys

	tmp := service.CreateCacheArgs{Name: name, Port: port, Retention: retention, Token: t}
	cache, err := serv.Create(tmp)
	if err != nil {
		zap.S().Errorf("Failed to create cache token, err: %+v", err)
		return err
	}

	cachePath, err := stor.CreateDir(name)
	if err != nil {
		zap.S().Errorf("Failed to create cache storage, err: %v", err)
		// TODO: clean dead entry to database
		return err
	}

	// Output for the user
	fmt.Printf("Binary Cache Created Successfully!\n")
	fmt.Printf("Name:      %s\n", cache.Name)
	fmt.Printf("Port:      %d\n", cache.Port)
	fmt.Printf("Token:     %s\n", cache.Token)
	fmt.Printf("Directory: %s", cachePath)
	if retention > 0 {
		fmt.Printf("Retention: %d days\n", cache.Retention)
	}

	return nil
}

func setup(cmd *cobra.Command, args []string) error {
	// Attempt to run parent's setup (e.g., root command)
	parent := cmd.Parent().Parent()
	if parent != nil && parent.PersistentPreRun != nil {
		zap.S().Debugf("Running parent setup %v ...", parent.Use)
		parent.PersistentPreRun(parent, args)
	}

	zap.S().Debug("Running workspace setup ...")
	return nil
}

// getServices gets the cache service
func getServices() (*service.CacheSrv, objstor.ObjectStorage) {
	var s *service.CacheSrv
	var storage objstor.ObjectStorage
	app.Invoke(func(serv *service.CacheSrv, objst objstor.ObjectStorage) {
		s = serv
		storage = objst
	})

	return s, storage
}

func remove(cmd *cobra.Command, args []string) error {
	name := args[0]

	zap.S().Debugf("Parsed args: %v %v %v", name)

	serv, _ := getServices()

	if err := serv.Delete(name); err != nil {
		zap.S().Errorf("Failed to create cache token, err: %+v", err)
		return err
	}

	// Output for the user
	fmt.Printf("Binary Cache Removed Successfully!\n")
	return nil
}
