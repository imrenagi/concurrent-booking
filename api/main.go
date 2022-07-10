package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/imrenagi/concurrent-booking/api/server"
)

func main() {
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
	api.Run(ctx, 9999)
}
