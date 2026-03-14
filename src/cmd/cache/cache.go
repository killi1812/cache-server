// Package cache
package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"github.com/killi1812/go-cache-server/util/objstor"
	"github.com/killi1812/go-cache-server/util/proc"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	serv         *service.CacheSrv
	stor         objstor.ObjectStorage
	ErrIsRunning = errors.New("err cache server is running")
)

const (
	_FOREGROUND_FLAG_NAME = "foreground"

	_RETENTION_FLAG_NAME = "retention"
	_NAME_FLAG_NAME      = "name"
	_ACCESS_FLAG_NAME    = "access"
	_PORT_FLAG_NAME      = "port"

	_PRIVATE_FLAG_NAME = "private"
	_PUBLIC_FLAG_NAME  = "public"
)

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
	cr.Flags().IntP(_RETENTION_FLAG_NAME, "r", -1, "Time to retain cache in weeks, 0 means forever")
	cr.Flags().StringP(_ACCESS_FLAG_NAME, "a", "private", "Set access to private/public")
	cr.RegisterFlagCompletionFunc(_ACCESS_FLAG_NAME, compleateAccess)

	st := &cobra.Command{
		Use:   "start [cache name]",
		Short: "start http server for specified cache",
		Args:  cobra.ExactArgs(1),
		RunE:  start,
	}
	st.Flags().BoolP(_FOREGROUND_FLAG_NAME, "f", false, "Run the app in foreground")

	up := &cobra.Command{
		Use:   "update [cache name]",
		Short: "update binary cache",
		Args:  cobra.ExactArgs(1),
		RunE:  update,
	}
	up.Flags().StringP(_NAME_FLAG_NAME, "n", "", "Change cache name to NAME")
	up.Flags().StringP(_ACCESS_FLAG_NAME, "a", "", "Change access to public/private")
	up.RegisterFlagCompletionFunc(_ACCESS_FLAG_NAME, compleateAccess)
	up.Flags().IntP(_PORT_FLAG_NAME, "p", 0, "Change cache port to PORT")
	up.Flags().IntP(_RETENTION_FLAG_NAME, "r", -1, "Change cache retention to RETENTION")

	ls := &cobra.Command{
		Use:   "list",
		Short: "list binary caches",
		Args:  cobra.NoArgs,
		RunE:  list,
	}
	ls.Flags().BoolP(_PUBLIC_FLAG_NAME, "P", false, "List public caches")
	ls.Flags().BoolP(_PRIVATE_FLAG_NAME, "p", false, "List private caches")
	ls.MarkFlagsMutuallyExclusive(_PUBLIC_FLAG_NAME, _PRIVATE_FLAG_NAME)

	ptr.AddCommand(cr, st, up, ls,
		&cobra.Command{
			Use:   "delete [cache name]",
			Short: "delete a binary cache",
			Args:  cobra.ExactArgs(1),
			RunE:  remove,
		},
		&cobra.Command{
			Use:   "info [cache name]",
			Short: "get info about binary cache",
			Args:  cobra.ExactArgs(1),
			RunE:  info,
		},

		&cobra.Command{
			Use:   "stop [cache name]",
			Short: "stop http server for specified cache",
			Args:  cobra.ExactArgs(1),
			RunE:  stop,
		},
	)

	return ptr
}

func cache(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func setup(cmd *cobra.Command, args []string) error {
	// Attempt to run parent's setup (e.g., root command)
	parent := cmd.Parent().Parent()
	if parent != nil && parent.PersistentPreRun != nil {
		zap.S().Debugf("Running parent setup %v ...", parent.Use)
		parent.PersistentPreRun(parent, args)
	}

	zap.S().Debug("Running workspace setup ...")

	app.Invoke(func(s *service.CacheSrv, storage objstor.ObjectStorage) {
		serv = s
		stor = storage
	})

	return nil
}

func create(cmd *cobra.Command, args []string) error {
	zap.S().Debugf("Trying to create binary cache ...")
	name := args[0]
	portstr := args[1]

	// TODO: add agruments for public/private

	zap.S().Debugf("Parsed args: %+v", args)

	port, err := strconv.Atoi(portstr)
	if err != nil {
		return fmt.Errorf("port is not a number %s", portstr)
	}

	retention, err := cmd.Flags().GetInt(_RETENTION_FLAG_NAME)
	if err != nil {
		zap.S().DPanicf("Failed to retrieve retention flag, err: %v", err)
	}

	t, err := auth.GenerateJwt(name)
	if err != nil {
		zap.S().Errorf("Failed to generate token, err: %v ", err)
		return err
	}

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
	fmt.Printf("Name:       %s\n", cache.Name)
	fmt.Printf("Port:       %d\n", cache.Port)
	fmt.Printf("Token:      %s\n", cache.Token)
	fmt.Printf("Directory:  %s\n", cachePath)
	if retention > 0 {
		fmt.Printf("Retention: %d weeks\n", cache.Retention)
	}

	return nil
}

// remove implements delete logic
func remove(cmd *cobra.Command, args []string) error {
	zap.S().Debugf("trying to delete binary cache ...")
	name := args[0]
	zap.S().Debugf("Parsed args: %v", name)

	err := serv.Delete(name)
	err2 := stor.DeleteFile(name)

	if err != nil {
		zap.S().Errorf("Failed to delete cache, err: %+v", err)
	}

	if err2 != nil {
		zap.S().Errorf("Failed to delete cache store, err: %+v", err2)
	}

	if err != nil && err2 != nil {
		return errors.Join(err, err2)
	}

	// Output for the user
	fmt.Printf("Binary Cache Removed Successfully!\n")
	return nil
}

func info(cmd *cobra.Command, args []string) error {
	zap.S().Debugf("trying to read info of binary cache ...")
	name := args[0]
	zap.S().Debugf("Parsed args: %v", name)

	cache, err := serv.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read cache , err: %+v", err)
		return err
	}

	zap.S().Debugf("Retrieved binary cache %s", name)
	tmpb := strings.Builder{}
	tmpe := json.NewEncoder(&tmpb)
	tmpe.SetIndent("", "   ")
	tmpe.Encode(cache)
	zap.S().Debug(tmpb.String())

	fmt.Printf("Name:      %s\n", cache.Name)
	fmt.Printf("Port:      %d\n", cache.Port)
	fmt.Printf("Access:    %s\n", cache.Access)
	fmt.Printf("Token:     %s\n", cache.Token)
	fmt.Printf("URL:       %s\n", cache.URL)
	if cache.Retention > 0 {
		fmt.Printf("Retention: %d weeks\n", cache.Retention)
	} else {
		fmt.Printf("Retention: indefinite\n")
	}

	return nil
}

func list(cmd *cobra.Command, args []string) error {
	zap.S().Infof("trying to list binary caches ...")

	// TODO: add json output
	caches, err := serv.ReadAll()
	if err != nil {
		zap.S().Errorf("Failed to create cache list, err: %+v", err)
		return err
	}

	zap.S().Debugf("Retrieved %d binary caches", len(caches))

	fmt.Printf("Found %d binary caches:\n", len(caches))
	for _, cache := range caches {
		fmt.Printf("\t%s\n", cache.Name)
	}

	return nil
}

func start(cmd *cobra.Command, args []string) error {
	zap.S().Infof("trying to start cache server ...")
	name := args[0]
	zap.S().Debugf("Parsed args: %v", name)

	foreground, err := cmd.Flags().GetBool(_FOREGROUND_FLAG_NAME)
	if err != nil {
		zap.S().DPanicf("Failed to retrieve foreground flag, err: %v", err)
	}

	cache, err := serv.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read cache , err: %+v", err)
		return err
	}

	addr := fmt.Sprintf("%s:%d", config.Config.CacheServer.Hostname, cache.Port)
	if foreground {
		zap.S().Infof("Starting server in foreground")
		app.Start(newApi(cache), addr)
	} else {
		zap.S().Infof("Starting server in backgound")
		err := proc.StartProcBackground(cache.Uuid.String() + ".pid")
		if err != nil {
			zap.S().Errorf("Failed to start cache '%s' server , err: %+v", name, err)
			return err
		}
		fmt.Printf("Cache Server '%s' Started:\t http://%s\n", name, addr)
	}

	return nil
}

func stop(cmd *cobra.Command, args []string) error {
	zap.S().Infof("trying to start cache server ...")
	name := args[0]
	zap.S().Debugf("Parsed args: %v", name)

	cache, err := serv.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read cache , err: %+v", err)
		return err
	}

	err = proc.StopProc(cache.Uuid.String() + ".pid")
	if err != nil {
		zap.S().Errorf("Failed to stop cache '%s' server , err: %+v", name, err)
		return err
	}

	fmt.Printf("Cache Server '%s' Stopped Successfully!\n", name)
	return nil
}

func update(cmd *cobra.Command, args []string) error {
	zap.S().Infof("trying to start cache server ...")
	name := args[0]
	zap.S().Debugf("Parsed args: %v", name)

	cache, err := serv.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read cache , err: %+v", err)
		return err
	}

	if proc.IsRunning(cache.Uuid.String() + ".pid") {
		zap.S().Errorf(" Cache '%s' server is running", cache.Name)
		return ErrIsRunning
	}

	retention, err := cmd.Flags().GetInt(_RETENTION_FLAG_NAME)
	if err != nil {
		zap.S().DPanicf("Failed to retrieve %s flag, err: %v", _RETENTION_FLAG_NAME, err)
	}
	newName, err := cmd.Flags().GetString(_NAME_FLAG_NAME)
	if err != nil {
		zap.S().DPanicf("Failed to retrieve %s flag, err: %v", _NAME_FLAG_NAME, err)
	}
	access, err := cmd.Flags().GetString(_ACCESS_FLAG_NAME)
	if err != nil {
		zap.S().DPanicf("Failed to retrieve %s flag, err: %v", _ACCESS_FLAG_NAME, err)
	}
	port, err := cmd.Flags().GetInt(_PORT_FLAG_NAME)
	if err != nil {
		zap.S().DPanicf("Failed to retrieve %s flag, err: %v", _PORT_FLAG_NAME, err)
	}

	t, err := auth.GenerateJwt(newName)
	if err != nil {
		zap.S().Errorf("Failed to generate token, err: %v ", err)
		return err
	}

	newCache := model.BinaryCache{
		Name:      newName,
		Retention: retention,
		Access:    model.ParseBinaryCacheAccess(access),
		Port:      port,
		Token:     t,
	}

	cache, err = serv.Update(name, newCache)
	if err != nil {
		zap.S().Errorf("Failed to update cache '%s', err: %v", name, err)
		return err
	}

	fmt.Printf("Name:      %s\n", cache.Name)
	fmt.Printf("Port:      %d\n", cache.Port)
	fmt.Printf("Access:    %s\n", cache.Access)
	fmt.Printf("Token:     %s\n", cache.Token)
	fmt.Printf("URL:       %s\n", cache.URL)
	if cache.Retention > 0 {
		fmt.Printf("Retention: %d weeks\n", cache.Retention)
	} else {
		fmt.Printf("Retention: indefinite\n")
	}

	return nil
}

func compleateAccess(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"public", "private"}, cobra.ShellCompDirectiveNoFileComp
}
