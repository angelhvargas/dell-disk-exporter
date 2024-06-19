package main

import (
	"log"
	"net/http"

	"github.com/angelhvargas/dell-disk-exporter/pkg/idrac"
	"github.com/angelhvargas/dell-disk-exporter/pkg/smart"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Start the Prometheus metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":9077", nil))
	}()

	// Initialize the IDRAC client with the default executor
	idracClient := idrac.NewClient(&idrac.DefaultCommandExecutor{})
	// Start the update loop
	go idracClient.UpdateMetrics()

	// Initialize the SMART metrics updater with the default executor
	smartMetrics := smart.NewMetrics(&smart.DefaultCommandExecutor{})
	go smartMetrics.UpdateMetrics()

	select {} // Block forever
}
