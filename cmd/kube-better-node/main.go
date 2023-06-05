package main

import (
	"flag"

	"github.com/decayofmind/kube-better-node/internal/controller"
	"github.com/decayofmind/kube-better-node/internal/k8s"

	"k8s.io/klog/v2"
)

var (
	dryRun    = flag.Bool("dry-run", false, "Dry run")
	tolerance = flag.Int("tolerance", 0, "Ignore certain weight difference")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	client, err := k8s.NewClient()
	if err != nil {
		panic(err.Error())
	}

	controller.Run(client, *dryRun, *tolerance)
}
