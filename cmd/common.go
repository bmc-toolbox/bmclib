package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/butler"

	"github.com/bmc-toolbox/bmcbutler/pkg/inventory"
	"github.com/bmc-toolbox/bmcbutler/pkg/metrics"
)

var (
	butlers        *butler.Butler
	commandWG      sync.WaitGroup
	metricsEmitter *metrics.Emitter
	stopChan       chan struct{}
	interrupt      bool
)

// post handles clean up actions
// - closes the butler channel
// - Waits for all go routines in commandWG to finish.
func post(butlerChan chan butler.Msg) {
	commandWG.Wait()
	if !interrupt {
		close(butlerChan)
		close(stopChan)
	}
	metricsEmitter.Close(true)
}

// Any flags to override configuration goes here.
func overrideConfigFromFlags() {
	if butlersToSpawn > 0 {
		runConfig.ButlersToSpawn = butlersToSpawn
	}

	if locations != "" {
		runConfig.Locations = strings.Split(locations, ",")
	}

	if resources != "" {
		runConfig.Resources = strings.Split(resources, ",")
	}

	if runConfig.DryRun {
		log.Info("Invoked with --dryrun.")
	}
}

// pre sets up required plumbing and returns two channels.
// - Spawn go routine to listen to interrupt signals
// - Setup metrics channel
// - Spawn the metrics forwarder go routine
// - Setup the inventory channel over which to receive assets
// - Based on the inventory source (dora/csv), Spawn the asset retriever go routine.
// - Spawn butlers
// - Return inventory channel, butler channel.
func pre() (inventoryChan chan []asset.Asset, butlerChan chan butler.Msg, stopChan chan struct{}) {

	overrideConfigFromFlags()

	//Channel used to indicate goroutines to exit.

	stopChan = make(chan struct{})

	//Initialize metrics collection.
	metricsEmitter = &metrics.Emitter{
		Config: runConfig,
		Logger: log,
	}

	metricsEmitter.Init()

	// A channel to receive inventory assets
	inventoryChan = make(chan []asset.Asset, 5)

	//determine inventory to fetch asset data.
	inventorySource := runConfig.InventoryParams.Source

	//based on inventory source, invoke assetRetriever
	var assetRetriever func()

	switch inventorySource {
	case "enc":
		inventoryInstance := inventory.Enc{
			Config:         runConfig,
			Log:            log,
			BatchSize:      10,
			AssetsChan:     inventoryChan,
			MetricsEmitter: metricsEmitter,
			StopChan:       stopChan,
		}

		assetRetriever = inventoryInstance.AssetRetrieve()
	case "csv":
		inventoryInstance := inventory.Csv{
			Config:     runConfig,
			Log:        log,
			AssetsChan: inventoryChan,
		}

		assetRetriever = inventoryInstance.AssetRetrieve()
	case "dora":
		inventoryInstance := inventory.Dora{
			Config:         runConfig,
			Log:            log,
			BatchSize:      10,
			AssetsChan:     inventoryChan,
			MetricsEmitter: metricsEmitter,
		}

		assetRetriever = inventoryInstance.AssetRetrieve()
	case "iplist":
		inventoryInstance := inventory.IPList{
			Channel:   inventoryChan,
			Config:    runConfig,
			BatchSize: 1,
			Log:       log,
		}

		assetRetriever = inventoryInstance.AssetRetrieve()
	default:
		fmt.Println("Unknown/no inventory source declared in cfg: ", inventorySource)
		os.Exit(1)
	}

	//invoke asset retriever routine
	//this routine returns assets over the inventoryChan.
	go assetRetriever()

	// Spawn butlers to work
	butlerChan = make(chan butler.Msg, 2)
	butlers = &butler.Butler{
		ButlerChan:     butlerChan,
		StopChan:       stopChan,
		Config:         runConfig,
		Log:            log,
		MetricsEmitter: metricsEmitter,
		SyncWG:         &commandWG,
	}

	go butlers.Runner()
	commandWG.Add(1)

	//setup a sigchan
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		interrupt = true
		log.Warn("Interrupt SIGINT/SIGTERM received.")
		close(butlerChan)
		close(stopChan)
	}()

	return inventoryChan, butlerChan, stopChan
}
