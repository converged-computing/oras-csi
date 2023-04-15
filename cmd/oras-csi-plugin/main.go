package main

import (
	"flag"

	"github.com/converged-computing/oras-csi/pkg/driver"
	"github.com/converged-computing/oras-csi/pkg/oras"
	log "github.com/sirupsen/logrus"
)

func main() {
	var (
		mode              = flag.String("mode", "unspecified", "")
		csiEndpoint       = flag.String("csi-endpoint", "unix:///var/lib/csi/sockets/pluginproxy/csi.sock", "CSI endpoint")
		nodeId            = flag.String("node-id", "", "")
		rootDir           = flag.String("root-dir", "/", "")
		pluginDataDir     = flag.String("plugin-data-dir", "/", "")
		handlersCount     = flag.Int("mount-points-count", 1, "")
		sanityTestRun     = flag.Bool("sanity-test-run", false, "")
		logLevel          = flag.Int("log-level", 5, "")
		orasLog           = flag.Bool("oras-logging", true, "")
		enforceNamespaces = flag.Bool("enforce-namespaces", true, "")
	)
	flag.Parse()

	// Setup logging variables
	driver.Init(*sanityTestRun, *logLevel, *orasLog)
	oras.Init(*logLevel)

	if *sanityTestRun {
		log.Infof("=============== SANITY TEST ===============")
	}
	log.Infof("Preparing artifact cache (mode: %s; node-id: %s; root-dir: %s; plugin-data-dir: %s enforce-namespaces: %s)",
		*mode, *nodeId, *rootDir, *pluginDataDir, *enforceNamespaces)

	var srv driver.Service
	var err error
	switch *mode {

	// Node service: run on the node where the plugin will be published
	case "node":
		srv, err = driver.NewNodeService(*rootDir, *pluginDataDir, *nodeId, *handlersCount, *enforceNamespaces)
		if err != nil {
			log.Error("main - couldn't create node service. Error: %s", err.Error())
			return
		}
	default:
		log.Error("main - unrecognized mode = %s", *mode)
		return
	}

	// This is the Identity service - "hello I am the ORAS csi!"
	if err = driver.StartService(&srv, *mode, *csiEndpoint); err != nil {
		log.Error("main - couldn't start service %s", err.Error())
	}
}
