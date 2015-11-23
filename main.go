package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/asteris-llc/pushmipullyu/dispatch"
	"github.com/asteris-llc/pushmipullyu/services/asana"
	"golang.org/x/net/context"
	"os"
	"os/signal"
	"time"
)

func main() {
	// logging
	logrus.SetLevel(logrus.DebugLevel)

	ctx, shutdown := context.WithCancel(context.Background())

	// dispatcher
	dispatch := dispatch.New()
	go dispatch.Run(ctx)

	// Asana
	asana := asana.New(os.Getenv("ASANA_TOKEN"))
	go asana.Handle(ctx, dispatch.Register("github"))

	defer shutdown()
	catch(shutdown)

	// give services time to finish and shut down
	time.Sleep(time.Second * 5)
}

func catch(handler func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for _ = range signals {
		logrus.Debug("received interrupt signal")
		handler()
		return
	}
}
