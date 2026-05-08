package rootcmd

import (
	"fmt"

	"github.com/killi1812/go-cache-server/util/auth"
	"github.com/spf13/cobra"
)

func NewUtilCmd() *cobra.Command {
	utilCmd := &cobra.Command{
		Use:   "util",
		Short: "Utility commands",
	}

	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Auth utility commands",
	}

	generateCmd := &cobra.Command{
		Use:   "generate [name]",
		Short: "Generate a JWT token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			token, err := auth.GenerateJwt(name)
			if err != nil {
				return err
			}
			fmt.Println(token)
			return nil
		},
	}

	authCmd.AddCommand(generateCmd)
	utilCmd.AddCommand(authCmd)

	return utilCmd
}
