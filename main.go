package main

import (
	"log"
	"net/http"

	"github.com/angelhvargas/dell-disk-exporter/pkg/idrac"
	"github.com/angelhvargas/dell-disk-exporter/pkg/smart"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Create a new Prometheus registry
	registry := prometheus.NewRegistry()

	// Register the default Prometheus collectors
	registry.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)

	// Start the Prometheus metrics server
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
		log.Fatal(http.ListenAndServe(":9077", nil))
	}()

	// Initialize the IDRAC client with the default executor and registry
	idracClient := idrac.NewClient(&idrac.DefaultCommandExecutor{}, registry)
	// Start the update loop
	go idracClient.UpdateMetrics()

	// Initialize the SMART metrics updater with the default executor and registry
	smartMetrics := smart.NewMetrics(&smart.DefaultCommandExecutor{}, registry)
	go smartMetrics.UpdateMetrics()

	select {} // Block forever
}
