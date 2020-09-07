package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jessevdk/go-flags"

	"github.com/thecoderstudio/apollo-agent/api"
	"github.com/thecoderstudio/apollo-agent/net"
)

var opts struct {
	Host              string `short:"h" long:"host" description:"Host address" required:"true"`
	AgentID           string `long:"agent-id" description:"Apollo agent id" required:"true"`
	Secret            string `long:"secret" description:"Apollo OAuth client secret" required:"true"`
	Shell             string `long:"shell" description:"Path to shell" default:"/bin/bash"`
	ReconnectInterval int    `short:"i" long:"reconnect-interval" description:"Interval in seconds between reconnection attempts" default:"5"`
}

func main() {
	log.SetFlags(0)
	parseArguments()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	host := net.GetHostFromURLString(opts.Host)
	if host == "" {
		log.Fatal("No valid host given")
	}
	middleware, err := api.CreateMiddleware(host, opts.AgentID, opts.Secret, opts.Shell, &interrupt)
	if err != nil {
		log.Fatal(err)
	}

	reconnectInterval := time.Duration(opts.ReconnectInterval) * time.Second

	err = middleware.Start(reconnectInterval)
	if err != nil {
		log.Fatal(err)
	}
}

func parseArguments() {
	_, err := flags.Parse(&opts)
	if err != nil {
		panic(err)
	}
}
