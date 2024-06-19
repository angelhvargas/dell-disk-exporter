package smart

import (
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"time"
)

type Metrics struct {
	CriticalWarning           int     `json:"critical_warning"`
	Temperature               float64 `json:"temperature"`
	AvailableSpare            float64 `json:"avail_spare"`
	SpareThreshold            float64 `json:"spare_thresh"`
	PercentageUsed            float64 `json:"percent_used"`
	DataUnitsRead             float64 `json:"data_units_read"`
	DataUnitsWritten          float64 `json:"data_units_written"`
	HostReadCommands          float64 `json:"host_read_commands"`
	HostWriteCommands         float64 `json:"host_write_commands"`
	ControllerBusyTime        float64 `json:"controller_busy_time"`
	PowerCycles               float64 `json:"power_cycles"`
	PowerOnHours              float64 `json:"power_on_hours"`
	UnsafeShutdowns           float64 `json:"unsafe_shutdowns"`
	MediaErrors               float64 `json:"media_errors"`
	NumErrorLogEntries        float64 `json:"num_err_log_entries"`
	WarningTemperatureTime    float64 `json:"warning_temp_time"`
	CriticalCompositeTempTime float64 `json:"critical_comp_time"`
	TemperatureSensor1        float64 `json:"temperature_sensor_1"`
	TemperatureSensor2        float64 `json:"temperature_sensor_2"`
	TemperatureSensor3        float64 `json:"temperature_sensor_3"`
	TemperatureSensor4        float64 `json:"temperature_sensor_4"`
	ThermalManagement1Trans   float64 `json:"thm_temp1_trans_count"`
	ThermalManagement2Trans   float64 `json:"thm_temp2_trans_count"`
	ThermalManagement1Time    float64 `json:"thm_temp1_total_time"`
	ThermalManagement2Time    float64 `json:"thm_temp2_total_time"`
	executor                  CommandExecutor
}

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

func NewMetrics(executor CommandExecutor) *Metrics {
	return &Metrics{executor: executor}
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
		drives, err := getNVMeDrives()
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
			// Update Prometheus metrics here
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
