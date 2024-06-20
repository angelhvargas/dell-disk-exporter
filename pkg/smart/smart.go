package smart

import (
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// CommandExecutor defines an interface for executing commands
type CommandExecutor interface {
	ExecuteCommand(name string, args ...string) ([]byte, error)
}

// DefaultCommandExecutor implements CommandExecutor
type DefaultCommandExecutor struct{}

func (e *DefaultCommandExecutor) ExecuteCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

// Function signature for detecting NVMe drives
type DetectDrivesFunc func() ([]string, error)

var detectNVMeDrives DetectDrivesFunc = getNVMeDrives

type Metrics struct {
	executor        CommandExecutor
	smartLogMetrics *prometheus.GaugeVec
	registry        *prometheus.Registry
}

func NewMetrics(executor CommandExecutor, registry *prometheus.Registry) *Metrics {
	smartLogMetrics := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nvme_smart_log",
			Help: "SMART log metrics for NVMe devices",
		},
		[]string{"device", "metric"},
	)
	registry.MustRegister(smartLogMetrics)
	return &Metrics{
		executor:        executor,
		smartLogMetrics: smartLogMetrics,
		registry:        registry,
	}
}

func (m *Metrics) GetSMARTLog(drive string) (map[string]interface{}, error) {
	output, err := m.executor.ExecuteCommand("nvme", "smart-log", "/dev/"+drive, "--output-format", "json")
	if err != nil {
		return nil, err
	}

	var smartLog map[string]interface{}
	if err := json.Unmarshal(output, &smartLog); err != nil {
		return parseNvmeSmartLogText(string(output))
	}

	return smartLog, nil
}

func parseNvmeSmartLogText(output string) (map[string]interface{}, error) {
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

	return smartLog, nil
}

func (m *Metrics) UpdateMetrics() {
	for {
		drives, err := detectNVMeDrives()
		if err != nil {
			log.Printf("Error detecting NVMe drives: %v", err)
			return
		}
		for _, drive := range drives {
			logData, err := m.GetSMARTLog(drive)
			if err != nil {
				log.Printf("Error getting SMART log for %s: %v", drive, err)
				continue
			}
			log.Printf("SMART Log for %s: %v", drive, logData)
			for key, value := range logData {
				floatValue, ok := value.(float64)
				if !ok {
					continue
				}
				m.smartLogMetrics.WithLabelValues(drive, key).Set(floatValue)
			}
		}
		time.Sleep(30 * time.Second) // Adjust the interval as needed
	}
}

func getNVMeDrives() ([]string, error) {
	cmd := exec.Command("lsblk", "-d", "-n", "-o", "NAME,TYPE")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var drives []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "nvme") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				drives = append(drives, parts[0])
			}
		}
	}
	return drives, nil
}
