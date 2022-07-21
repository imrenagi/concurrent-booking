package commands

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/imrenagi/concurrent-booking/worker"
)

func workerCmd() *cobra.Command {

	var command = &cobra.Command{
		Use:   "worker",
		Short: "Run the worker",
		Run: func(c *cobra.Command, args []string) {
			worker := worker.NewWorker()

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
			worker.Run(ctx)
		},
	}

	return command

}
