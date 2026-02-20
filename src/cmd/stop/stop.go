// Package stop
package stop

import (
	"errors"
	"fmt"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/util/proc"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var noAsk = false

var ErrFailedToStop = errors.New("failed stopping the cache-server")

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:   "stop",
		Short: "Stop cache server",
		Long:  `Stop cache server running in the background`,
		RunE:  stop,
	}

	ptr.PersistentFlags().BoolVarP(&noAsk, "no-ask", "n", false, "don't ask questions assume default answer for all")

	return ptr
}

func stop(cmd *cobra.Command, args []string) error {
	err := proc.StopProc(app.PID_FILE_NAME, proc.StopOpts{NoAsk: noAsk, Search: true})
	if err != nil {
		zap.S().Errorf("Failed to stop the proccess, err: %v", err)
		return err
	}

	fmt.Println("Server Stopped Successfully!")
	return nil
}
