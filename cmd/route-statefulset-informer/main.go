package main

import (
	"os"

	cmdcommons "code.cloudfoundry.org/eirini/cmd"
	k8sroute "code.cloudfoundry.org/eirini/k8s/informers/route"
	"code.cloudfoundry.org/eirini/k8s/informers/route/event"
	"code.cloudfoundry.org/eirini/route"
	"code.cloudfoundry.org/lager"
	"github.com/jessevdk/go-flags"
)

type options struct {
	ConfigFile string `short:"c" long:"config" description:"Config for running route-collector" required:"true"`
}

func main() {
	var opts options
	_, err := flags.ParseArgs(&opts, os.Args)
	cmdcommons.ExitWithError(err)

	cfg, err := route.ReadConfig(opts.ConfigFile)
	cmdcommons.ExitWithError(err)

	logger := lager.NewLogger("route-statefulset-informer")
	logger.RegisterSink(lager.NewPrettySink(os.Stdout, lager.DEBUG))

	routeEmitter, err := route.NewEmitterFromConfig(cfg.NatsIP, cfg.NatsPort, cfg.NatsPassword, logger)
	cmdcommons.ExitWithError(err)

	clientset := cmdcommons.CreateKubeClient(cfg.ConfigPath)

	deleteHandler := event.StatefulSetDeleteHandler{
		Pods:         clientset.CoreV1().Pods(cfg.Namespace),
		Logger:       logger.Session("uri-delete-informer"),
		RouteEmitter: routeEmitter,
	}

	updateHandler := event.URIAnnotationUpdateHandler{
		Pods:         clientset.CoreV1().Pods(cfg.Namespace),
		Logger:       logger.Session("update-handler"),
		RouteEmitter: routeEmitter,
	}

	uriInformer := k8sroute.NewURIChangeInformer(
		clientset,
		cfg.Namespace,
		updateHandler,
		deleteHandler,
	)

	uriInformer.Start()
}
