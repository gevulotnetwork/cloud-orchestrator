package main

import (
	"flag"

	"github.com/gevulotnetwork/cloud-orchestrator/cloud"
	"github.com/gevulotnetwork/cloud-orchestrator/config"
	"github.com/gevulotnetwork/cloud-orchestrator/server"
)

func main() {
	var listenAddr = flag.String("listen-addr", "localhost:0", "gRPC service listen address")
	flag.Parse()

	cfgFactory := config.NewFactory()

	orchestrator := cloud.NewOrchestrator(cfgFactory)
	server.Start(server.Config{
		ListenAddr:   *listenAddr,
		Orchestrator: orchestrator,
	})
}
