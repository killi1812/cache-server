// Package storepath
package storepath

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/service"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serv *service.StorePathSrv

/*
	 NewCmd create a new store-path cmd
				list              List store paths
		    delete            Delete store path
		    info              Display info about store path
*/
func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:               "store-path",
		Short:             "Manage deployment paths",
		Run:               storepath,
		PersistentPreRunE: setup,
	}

	ptr.AddCommand(
		&cobra.Command{
			Use:   "list [cache name]",
			Short: "List store paths in a specific cache",
			Args:  cobra.ExactArgs(1),
			RunE:  list,
		},
		&cobra.Command{
			Use:   "delete [store hash] [cache name]",
			Short: "",
			Args:  cobra.ExactArgs(2),
			RunE:  remove,
		},
		&cobra.Command{
			Use:   "info [store hash] [cache name]",
			Short: "",
			Args:  cobra.ExactArgs(2),
			RunE:  info,
		},
	)

	return ptr
}

func storepath(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func list(cmd *cobra.Command, args []string) error {
	zap.S().Infof("Trying to list store paths ...")
	cache := args[0]
	zap.S().Debugf("Parsed args %+v", args)

	paths, err := serv.ReadAll(cache)
	if err != nil {
		zap.S().Errorf("Failed to read store paths for cache '%s', err: %v", cache, err)
	}

	fmt.Printf("Store paths for cache '%s':\n", cache)
	for _, path := range paths {
		fmt.Printf("\t%s\n", path.StoreHash)
	}
	return nil
}

func info(cmd *cobra.Command, args []string) error {
	zap.S().Infof("Trying to read info for store path ...")
	storeHash := args[0]
	cache := args[1]
	zap.S().Debugf("Parsed args %+v", args)

	path, err := serv.Read(storeHash, cache)
	if err != nil {
		zap.S().Errorf("Failed to read store path '%s' for cache '%s', err: %v", storeHash, cache, err)
	}

	zap.S().Debugf("Retrieved workspace %s", cache)
	tmpb := strings.Builder{}
	tmpe := json.NewEncoder(&tmpb)
	tmpe.SetIndent("", "   ")
	tmpe.Encode(path)
	zap.S().Debug(tmpb.String())

	fmt.Printf("Hash:       %v\n", path.StoreHash)
	fmt.Printf("Suffix:     %v\n", path.StoreSuffix)
	fmt.Printf("File Hash:  %v\n", path.FileHash)
	fmt.Printf("File Size:  %v\n", path.FileSize)
	fmt.Printf("Nar Hash:   %v\n", path.NarHash)
	fmt.Printf("Nar Size:   %v\n", path.NarSize)
	fmt.Printf("Deriver:    %v\n", path.Deriver)
	fmt.Printf("References: %v\n", path.References)
	fmt.Printf("Cache:      %v\n", path.BinaryCache.Name)

	return nil
}

func remove(cmd *cobra.Command, args []string) error {
	zap.S().Infof("Trying to remove store path ...")
	storeHash := args[0]
	cache := args[1]
	zap.S().Debugf("Parsed args %+v", args)

	err := serv.Delete(storeHash, cache)
	if err != nil {
		zap.S().Errorf("Failed to delete store path '%s' for cache '%s', err: %v", storeHash, cache, err)
		return err
	}

	fmt.Printf("Store Path Removed Successfully!\n")
	return nil
}

// setup for agent subcommands
func setup(cmd *cobra.Command, args []string) error {
	// run parent setup

	parent := cmd.Parent()
	if parent != nil && parent.PersistentPreRun != nil {
		zap.S().Debugf("Running parent setup %d ...", parent.Use)
		parent.PersistentPreRun(parent, args)
	}

	zap.S().Debugf("Running agent setup...")
	app.Invoke(func(s *service.StorePathSrv) {
		serv = s
	})

	return nil
}
