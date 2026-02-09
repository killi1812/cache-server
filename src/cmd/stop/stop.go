// Package stop
package stop

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/killi1812/go-cache-server/util/pid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var noAsk = false

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:   "stop",
		Short: "Stop cache server",
		Long:  `Stop cache server running in the background`,
		Run:   stop,
	}

	ptr.PersistentFlags().BoolVarP(&noAsk, "no-ask", "n", false, "don't ask questions assume default answer for all")

	return ptr
}

func stop(cmd *cobra.Command, args []string) {
	procPid := -1
	// check for .pid file
	if !pid.CheckPid() {
		zap.S().Errorf("No pid file")

		// check for cache-server process and ask to stop it
		p, err := pid.FindPidByName()
		if err != nil {
			zap.S().Errorf("No process with name cache-server is running")
			return
		}
		zap.S().Infof("Found a cache-server proceess with pid %d ", p)
		if !noAsk {
			fmt.Print("do you want to stop it? [Y/n]: ")
			scanner := bufio.NewScanner(os.Stdin)

			if scanner.Scan() {
				response := strings.ToLower(strings.TrimSpace(scanner.Text()))
				if response == "n" || response == "no" {
					return
				}
			}
		}
		// set pid to found process pid
		procPid = p
	} else {
		p, err := pid.ReadPid()
		if err != nil {
			zap.S().Errorf("Failed stopping the cache-server")
			return
		}

		// set pid to .pid file value
		procPid = p
	}
	fmt.Printf("procPid: %v\n", procPid)
}
