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
		log.Fatal(http.ListenAndServe(":8000", nil))
	}()

	// Initialize the IDRAC client
	idracClient := idrac.NewClient()
	// Start the update loop
	idracClient.UpdateMetrics()

	// Initialize the SMART metrics updater
	smartMetrics := smart.NewMetrics()
	smartMetrics.UpdateMetrics()

	select {} // Block forever
}
