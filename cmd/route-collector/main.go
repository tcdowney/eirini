package main

import (
	"os"
	"time"

	cmdcommons "code.cloudfoundry.org/eirini/cmd"
	"code.cloudfoundry.org/eirini/k8s"
	"code.cloudfoundry.org/eirini/route"
	"code.cloudfoundry.org/eirini/util"
	"code.cloudfoundry.org/lager"
	"github.com/jessevdk/go-flags"
)

type options struct {
	ConfigFile string `short:"c" long:"config" description:"Config for running route-collector"`
}

func main() {
	var opts options
	_, err := flags.ParseArgs(&opts, os.Args)
	cmdcommons.ExitWithError(err)

	cfg, err := route.ReadConfig(opts.ConfigFile)
	cmdcommons.ExitWithError(err)

	logger := lager.NewLogger("route-collector")
	logger.RegisterSink(lager.NewPrettySink(os.Stdout, lager.DEBUG))

	routeEmitter, err := route.NewEmitterFromConfig(cfg.NatsIP, cfg.NatsPort, cfg.NatsPassword, logger)
	cmdcommons.ExitWithError(err)

	clientset := cmdcommons.CreateKubeClient(cfg.ConfigPath)
	collector := k8s.NewRouteCollector(clientset, cfg.Namespace, logger)

	scheduler := route.CollectorScheduler{
		Collector: collector,
		Emitter:   routeEmitter,
		Scheduler: &util.TickerTaskScheduler{
			Ticker: time.NewTicker(30 * time.Second),
			Logger: logger.Session("scheduler"),
		},
	}
	scheduler.Start()
}
