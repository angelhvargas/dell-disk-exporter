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

type Metrics struct {
	executor        CommandExecutor
	smartLogMetrics *prometheus.GaugeVec
	nvmePresence    *prometheus.GaugeVec
	absentDrives    map[string]time.Time
	absentDuration  time.Duration
}

// Exported for testing
var GetNVMeDrives = func() ([]string, error) {
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

func NewMetrics(executor CommandExecutor, registry *prometheus.Registry, absentDuration time.Duration) *Metrics {
	smartLogMetrics := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nvme_smart_log",
			Help: "SMART log metrics for NVMe devices",
		},
		[]string{"device", "metric"},
	)
	nvmePresence := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nvme_presence",
			Help: "Presence of NVMe devices",
		},
		[]string{"device"},
	)
	registry.MustRegister(smartLogMetrics)
	registry.MustRegister(nvmePresence)
	return &Metrics{
		executor:        executor,
		smartLogMetrics: smartLogMetrics,
		nvmePresence:    nvmePresence,
		absentDrives:    make(map[string]time.Time),
		absentDuration:  absentDuration,
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
		drives, err := GetNVMeDrives()
		if err != nil {
			log.Printf("Error detecting NVMe drives: %v", err)
			return
		}

		currentDrives := make(map[string]bool)
		for _, drive := range drives {
			currentDrives[drive] = true
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
			m.nvmePresence.WithLabelValues(drive).Set(1)
			delete(m.absentDrives, drive)
		}

		for drive, timestamp := range m.absentDrives {
			if !currentDrives[drive] {
				if time.Since(timestamp) > m.absentDuration {
					m.nvmePresence.DeleteLabelValues(drive)
					m.smartLogMetrics.DeleteLabelValues(drive)
				} else {
					m.nvmePresence.WithLabelValues(drive).Set(0)
				}
			}
		}

		for drive := range currentDrives {
			if _, found := m.absentDrives[drive]; !found {
				m.absentDrives[drive] = time.Now()
			}
		}

		time.Sleep(30 * time.Second) // Adjust the interval as needed
	}
}
