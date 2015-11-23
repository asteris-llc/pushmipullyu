package main

import (
	"golang.org/x/net/context"
	"os"
	"os/signal"
)

func main() {
	ctx, shutdown := context.WithCancel(context.Background())

	// dispatcher
	dispatch := NewDispatch()
	dispatch.Run(ctx)

	// TODO: services

	defer shutdown()
}

func catch(handler func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for _ = range signals {
		handler()
	}
}
