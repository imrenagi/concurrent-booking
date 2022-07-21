package commands

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/imrenagi/concurrent-booking/server"
)

func serverCmd() *cobra.Command {
	var (
		listenPort int
	)
	var command = &cobra.Command{
		Use:   "server",
		Short: "Run the API server",
		Long:  "Run the API server",
		Run: func(c *cobra.Command, args []string) {
			api := server.NewServer()

			ctx := context.Background()
			ctx, cancel := context.WithCancel(ctx)

			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt)
			signal.Notify(ch, syscall.SIGTERM)

			go func() {
				oscall := <-ch
				log.Debug().Msgf("system call:%+v", oscall)
				cancel()
			}()

			// TODO Move shutdown handling inside object and remove context parameter.
			api.Run(ctx, listenPort)
		},
	}

	command.Flags().IntVar(&listenPort, "port", 9999, "Listen on given port")
	return command

}
