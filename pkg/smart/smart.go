package smart

import (
	"encoding/json"
	"log"
	"os/exec"
	"strings"
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
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) GetSMARTLog(drive string) (map[string]interface{}, error) {
	cmd := exec.Command("nvme", "smart-log", "/dev/"+drive, "--output-format", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var logData map[string]interface{}
	if err := json.Unmarshal(output, &logData); err != nil {
		return nil, err
	}
	return logData, nil
}

func (m *Metrics) UpdateMetrics() {
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
