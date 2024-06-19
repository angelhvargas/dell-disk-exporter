package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	raidStatusGauges    = make(map[string]prometheus.Gauge)
	raidRedundancyGauges = make(map[string]prometheus.Gauge)
	raidSizeGauges      = make(map[string]prometheus.Gauge)
	raidLayoutGauges    = make(map[string]prometheus.Gauge)

	nvmeHealthGauges    = make(map[string]prometheus.Gauge)
	nvmePresenceGauges  = make(map[string]prometheus.Gauge)
	nvmeSmartLogGauges  = make(map[string]prometheus.Gauge)
)

func getRaidStatus() map[string]map[string]string {
	cmd := exec.Command("racadm", "raid", "get", "vdisks", "-o", "-p", "layout,status,RemainingRedundancy,Size")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing racadm command:", err)
        fmt.Println("Command output:", string(output))
		return nil
	}

	lines := strings.Split(string(output), "\n")
	raidStatuses := make(map[string]map[string]string)
	var currentVdisk string

	for _, line := range lines {
		if strings.HasPrefix(line, "Disk.Virtual") {
			currentVdisk = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if currentVdisk != "" && strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if _, exists := raidStatuses[currentVdisk]; !exists {
				raidStatuses[currentVdisk] = make(map[string]string)
			}
			raidStatuses[currentVdisk][key] = value
		}
	}

	return raidStatuses
}

func detectNvmeDrives() []string {
	cmd := exec.Command("lsblk", "-d", "-n", "-o", "NAME,TYPE")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing lsblk command:", err)
		return nil
	}

	var nvmeDrives []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == "disk" && strings.HasPrefix(parts[0], "nvme") {
			nvmeDrives = append(nvmeDrives, parts[0])
		}
	}
	return nvmeDrives
}

func getNvmeSmartLog(nvmeDrive string) map[string]interface{} {
	cmd := exec.Command("nvme", "smart-log", fmt.Sprintf("/dev/%s", nvmeDrive), "--output-format", "json")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing nvme smart-log command:", err)
		return nil
	}

	var smartLog map[string]interface{}
	if err := json.Unmarshal(output, &smartLog); err != nil {
		return parseNvmeSmartLogText(string(output))
	}

	return smartLog
}

func parseNvmeSmartLogText(output string) map[string]interface{} {
	smartLog := make(map[string]interface{})
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			smartLog[key] = value
		}
	}

	return smartLog
}

func convertTemperature(value string) float64 {
	re := regexp.MustCompile(`(\d+)\s*C\s*\(\d+\s*Kelvin\)`)
	match := re.FindStringSubmatch(value)
	if match != nil {
		if temp, err := strconv.ParseFloat(match[1], 64); err == nil {
			return temp
		}
	}

	if temp, err := strconv.ParseFloat(value, 64); err == nil {
		return temp - 273.15
	}
	return 0
}

func updateMetrics() {
	var previousNvmeDrives []string

	for {
		// Update RAID status
		raidStatuses := getRaidStatus()
		for vdisk, metrics := range raidStatuses {
			if _, exists := raidStatusGauges[vdisk]; !exists {
				raidStatusGauges[vdisk] = prometheus.NewGauge(prometheus.GaugeOpts{
					Name: fmt.Sprintf("raid_status_%s", strings.ReplaceAll(strings.ReplaceAll(vdisk, ".", "_"), "-", "_")),
					Help: fmt.Sprintf("Status of virtual disk %s", vdisk),
				})
				raidRedundancyGauges[vdisk] = prometheus.NewGauge(prometheus.GaugeOpts{
					Name: fmt.Sprintf("raid_redundancy_%s", strings.ReplaceAll(strings.ReplaceAll(vdisk, ".", "_"), "-", "_")),
					Help: fmt.Sprintf("Remaining Redundancy of virtual disk %s", vdisk),
				})
				raidSizeGauges[vdisk] = prometheus.NewGauge(prometheus.GaugeOpts{
					Name: fmt.Sprintf("raid_size_%s", strings.ReplaceAll(strings.ReplaceAll(vdisk, ".", "_"), "-", "_")),
					Help: fmt.Sprintf("Size of virtual disk %s", vdisk),
				})
				raidLayoutGauges[vdisk] = prometheus.NewGauge(prometheus.GaugeOpts{
					Name: fmt.Sprintf("raid_layout_%s", strings.ReplaceAll(strings.ReplaceAll(vdisk, ".", "_"), "-", "_")),
					Help: fmt.Sprintf("Layout of virtual disk %s", vdisk),
				})
				prometheus.MustRegister(raidStatusGauges[vdisk], raidRedundancyGauges[vdisk], raidSizeGauges[vdisk], raidLayoutGauges[vdisk])
			}

			status := 0.0
			if metrics["Status"] == "Ok" {
				status = 1.0
			}
			raidStatusGauges[vdisk].Set(status)

			redundancy, _ := strconv.ParseFloat(metrics["RemainingRedundancy"], 64)
			raidRedundancyGauges[vdisk].Set(redundancy)

			size := parseSize(metrics["Size"])
			raidSizeGauges[vdisk].Set(size)

			layout := 0.0
			if metrics["Layout"] != "" {
				layout = 1.0
			}
			raidLayoutGauges[vdisk].Set(layout)
		}

		// Detect NVMe drives
		currentNvmeDrives := detectNvmeDrives()

		// Update NVMe presence metrics
		for _, nvmeDrive := range previousNvmeDrives {
			if !contains(currentNvmeDrives, nvmeDrive) {
				if _, exists := nvmePresenceGauges[nvmeDrive]; !exists {
					nvmePresenceGauges[nvmeDrive] = prometheus.NewGauge(prometheus.GaugeOpts{
						Name: fmt.Sprintf("nvme_presence_%s", nvmeDrive),
						Help: fmt.Sprintf("Presence of NVMe device %s", nvmeDrive),
					})
					prometheus.MustRegister(nvmePresenceGauges[nvmeDrive])
				}
				nvmePresenceGauges[nvmeDrive].Set(0)
			}
		}

		for _, nvmeDrive := range currentNvmeDrives {
			if _, exists := nvmePresenceGauges[nvmeDrive]; !exists {
				nvmePresenceGauges[nvmeDrive] = prometheus.NewGauge(prometheus.GaugeOpts{
					Name: fmt.Sprintf("nvme_presence_%s", nvmeDrive),
					Help: fmt.Sprintf("Presence of NVMe device %s", nvmeDrive),
				})
				prometheus.MustRegister(nvmePresenceGauges[nvmeDrive])
			}
			nvmePresenceGauges[nvmeDrive].Set(1)

			if _, exists := nvmeHealthGauges[nvmeDrive]; !exists {
				nvmeHealthGauges[nvmeDrive] = prometheus.NewGauge(prometheus.GaugeOpts{
					Name: fmt.Sprintf("nvme_health_%s", nvmeDrive),
					Help: fmt.Sprintf("Health of NVMe device %s", nvmeDrive),
				})
				prometheus.MustRegister(nvmeHealthGauges[nvmeDrive])
			}

			// Get NVMe SMART log
			smartLog := getNvmeSmartLog(nvmeDrive)

			// Update health status
			health := 0.0
			if smartLog["critical_warning"] == 0 {
				health = 1.0
			}
			nvmeHealthGauges[nvmeDrive].Set(health)

			// Update SMART log metrics
			for key, value := range smartLog {
				gaugeKey := fmt.Sprintf("nvme_%s_%s", nvmeDrive, strings.ReplaceAll(strings.ToLower(strings.ReplaceAll(key, " ", "_")), "%", "percent"))
				if _, exists := nvmeSmartLogGauges[gaugeKey]; !exists {
					nvmeSmartLogGauges[gaugeKey] = prometheus.NewGauge(prometheus.GaugeOpts{
						Name: gaugeKey,
						Help: fmt.Sprintf("%s for NVMe device %s", key, nvmeDrive),
					})
					prometheus.MustRegister(nvmeSmartLogGauges[gaugeKey])
				}

				switch v := value.(type) {
				case string:
					if strings.Contains(key, "temperature") || strings.Contains(key, "temp") {
						nvmeSmartLogGauges[gaugeKey].Set(convertTemperature(v))
					} else {
						if numValue, err := strconv.ParseFloat(strings.Fields(v)[0], 64); err == nil {
							nvmeSmartLogGauges[gaugeKey].Set(numValue)
						}
					}
				case float64:
					nvmeSmartLogGauges[gaugeKey].Set(v)
				}
			}
		}

		previousNvmeDrives = currentNvmeDrives
		time.Sleep(30 * time.Second)
	}
}

func parseSize(size string) float64 {
	re := regexp.MustCompile(`[\d\.]+`)
	matches := re.FindString(size)
	if matches != "" {
		if sizeGb, err := strconv.ParseFloat(matches, 64); err == nil {
			return sizeGb
		}
	}
	return 0
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func main() {
	// Start the Prometheus metrics server
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":8000", nil)
	}()

	// Start the metrics update thread
	go updateMetrics()

	// Keep the main thread alive
	select {}
}
