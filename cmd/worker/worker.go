package main

import (
	"context"
	// "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/imrenagi/concurrent-booking/worker"
)

func main() {
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
}
