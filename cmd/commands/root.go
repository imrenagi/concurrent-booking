package commands

import (
	"github.com/spf13/cobra"
)

var (
	cliName = "booking"
)

// NewRootCommand returns a new instance of an command
func NewRootCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:   cliName,
		Short: "Run service",
		Run: func(c *cobra.Command, args []string) {
			c.HelpFunc()(c, args)
		},
	}
	command.AddCommand(serverCmd(), workerCmd())
	return command
}
