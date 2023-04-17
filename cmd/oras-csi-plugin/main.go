package main

import (
	"flag"

	"github.com/converged-computing/oras-csi/pkg/driver"
	"github.com/converged-computing/oras-csi/pkg/oras"
	log "github.com/sirupsen/logrus"
)

func main() {
	var (
		mode              = flag.String("mode", driver.DefaultMode, "")
		csiEndpoint       = flag.String("csi-endpoint", driver.DefaultCSISocket, "CSI endpoint")
		nodeId            = flag.String("node-id", driver.DefaultNodeID, "")
		rootDir           = flag.String("root-dir", driver.DefaultRootDir, "")
		pluginDataDir     = flag.String("plugin-data-dir", driver.DefaultPluginDataDir, "")
		handlersCount     = flag.Int("handlers-count", driver.DefaultHandlersCount, "")
		sanityTestRun     = flag.Bool("sanity-test-run", driver.DefaultSanityTestRun, "")
		logLevel          = flag.Int("log-level", driver.DefaultLogLevel, "")
		orasLog           = flag.Bool("oras-logging", oras.DefaultEnableLogging, "")
		enforceNamespaces = flag.Bool("enforce-namespaces", driver.DefaultEnforceNamespaces, "")
	)
	flag.Parse()

	// Setup logging variables
	driver.Init(*sanityTestRun, *logLevel, *orasLog)
	oras.Init(*logLevel)

	if *sanityTestRun {
		log.Infof("<********** Sanity Test Run **********>")
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
