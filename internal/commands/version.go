package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.0.1"

// @ version cmd
func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Displays the version of the secret tool",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("secret version %s\n", Version)
		},
	}
	return cmd
}
