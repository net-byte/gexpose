package main

import (
	"flag"

	"github.com/net-byte/gexpose/client"
	"github.com/net-byte/gexpose/config"
	"github.com/net-byte/gexpose/server"
)

func main() {
	config := config.Config{}
	flag.StringVar(&config.LocalAddr, "l", ":9000", "local address")
	flag.StringVar(&config.ServerAddr, "s", ":8701", "server address")
	flag.StringVar(&config.ProxyAddr, "p", ":8702", "proxy address")
	flag.StringVar(&config.ExposeAddr, "e", ":8703", "expose address")
	flag.StringVar(&config.Key, "k", "Xn2r4u7x!A%D*G8", "encryption key")
	flag.BoolVar(&config.ServerMode, "server", false, "server mode")
	flag.IntVar(&config.Timeout, "t", 30, "dial timeout in seconds")
	flag.Parse()
	if config.ServerMode {
		server.Start(config)
	} else {
		client.Start(config)
	}
}
