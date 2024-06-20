package smart

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// MockCommandExecutor implements CommandExecutor for testing
type MockCommandExecutor struct {
	MockOutput string
	MockError  error
}

func (e *MockCommandExecutor) ExecuteCommand(name string, args ...string) ([]byte, error) {
	if e.MockError != nil {
		return nil, e.MockError
	}
	return []byte(e.MockOutput), nil
}

// mockGetNVMeDrives simulates the function to detect NVMe drives for testing
var mockGetNVMeDrives = func() ([]string, error) {
	return []string{"nvme0n1"}, nil
}

// mockGetNVMeDrivesAbsent simulates the function to detect NVMe drives, but simulates the drive becoming absent
var mockGetNVMeDrivesAbsent = func() ([]string, error) {
	return []string{}, nil
}

func TestGetSMARTLog(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		MockOutput: `{
  "critical_warning" : 0,
  "temperature" : 301,
  "avail_spare" : 100,
  "spare_thresh" : 5,
  "percent_used" : 15,
  "data_units_read" : 499296134,
  "data_units_written" : 1474968593,
  "host_read_commands" : 9347931143,
  "host_write_commands" : 51493840602,
  "controller_busy_time" : 1974546,
  "power_cycles" : 290,
  "power_on_hours" : 38313,
  "unsafe_shutdowns" : 139,
  "media_errors" : 0,
  "num_err_log_entries" : 17,
  "warning_temp_time" : 0,
  "critical_comp_time" : 0,
  "temperature_sensor_1" : 306,
  "temperature_sensor_2" : 301,
  "temperature_sensor_3" : 296,
  "temperature_sensor_4" : 295,
  "thm_temp1_trans_count" : 0,
  "thm_temp2_trans_count" : 0,
  "thm_temp1_total_time" : 0,
  "thm_temp2_total_time" : 0
}`,
	}

	registry := prometheus.NewRegistry()
	metrics := NewMetrics(mockExecutor, registry)
	log, err := metrics.GetSMARTLog("nvme0n1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(log) == 0 {
		t.Fatalf("Expected SMART log, got none")
	}
	if log["temperature"].(float64) != 301 {
		t.Fatalf("Expected temperature to be 301, got %v", log["temperature"])
	}
}

func TestGetSMARTLogError(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		MockError: errors.New("command error"),
	}

	registry := prometheus.NewRegistry()
	metrics := NewMetrics(mockExecutor, registry)
	_, err := metrics.GetSMARTLog("nvme0n1")
	if err == nil {
		t.Fatalf("Expected error, got none")
	}
}

func TestUpdateMetrics(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		MockOutput: `{
  "critical_warning" : 0,
  "temperature" : 301,
  "avail_spare" : 100,
  "spare_thresh" : 5,
  "percent_used" : 15,
  "data_units_read" : 499296134,
  "data_units_written" : 1474968593,
  "host_read_commands" : 9347931143,
  "host_write_commands" : 51493840602,
  "controller_busy_time" : 1974546,
  "power_cycles" : 290,
  "power_on_hours" : 38313,
  "unsafe_shutdowns" : 139,
  "media_errors" : 0,
  "num_err_log_entries" : 17,
  "warning_temp_time" : 0,
  "critical_comp_time" : 0,
  "temperature_sensor_1" : 306,
  "temperature_sensor_2" : 301,
  "temperature_sensor_3" : 296,
  "temperature_sensor_4" : 295,
  "thm_temp1_trans_count" : 0,
  "thm_temp2_trans_count" : 0,
  "thm_temp1_total_time" : 0,
  "thm_temp2_total_time" : 0
}`,
	}

	// Mock getNVMeDrives function by setting it to the mock function
	originalGetNVMeDrives := getNVMeDrives
	getNVMeDrives = mockGetNVMeDrives
	defer func() { getNVMeDrives = originalGetNVMeDrives }()

	registry := prometheus.NewRegistry()
	metrics := NewMetrics(mockExecutor, registry)

	go metrics.UpdateMetrics()

	// Allow some time for metrics to be updated
	time.Sleep(1 * time.Second)

	// Test SMART log metrics
	expectedMetrics := `
# HELP nvme_smart_log SMART log metrics for NVMe devices
# TYPE nvme_smart_log gauge
nvme_smart_log{device="nvme0n1",metric="avail_spare"} 100
nvme_smart_log{device="nvme0n1",metric="controller_busy_time"} 1.974546e+06
nvme_smart_log{device="nvme0n1",metric="critical_comp_time"} 0
nvme_smart_log{device="nvme0n1",metric="critical_warning"} 0
nvme_smart_log{device="nvme0n1",metric="data_units_read"} 4.99296134e+08
nvme_smart_log{device="nvme0n1",metric="data_units_written"} 1.474968593e+09
nvme_smart_log{device="nvme0n1",metric="host_read_commands"} 9.347931143e+09
nvme_smart_log{device="nvme0n1",metric="host_write_commands"} 5.1493840602e+10
nvme_smart_log{device="nvme0n1",metric="media_errors"} 0
nvme_smart_log{device="nvme0n1",metric="num_err_log_entries"} 17
nvme_smart_log{device="nvme0n1",metric="percent_used"} 15
nvme_smart_log{device="nvme0n1",metric="power_cycles"} 290
nvme_smart_log{device="nvme0n1",metric="power_on_hours"} 38313
nvme_smart_log{device="nvme0n1",metric="spare_thresh"} 5
nvme_smart_log{device="nvme0n1",metric="temperature"} 301
nvme_smart_log{device="nvme0n1",metric="temperature_sensor_1"} 306
nvme_smart_log{device="nvme0n1",metric="temperature_sensor_2"} 301
nvme_smart_log{device="nvme0n1",metric="temperature_sensor_3"} 296
nvme_smart_log{device="nvme0n1",metric="temperature_sensor_4"} 295
nvme_smart_log{device="nvme0n1",metric="thm_temp1_total_time"} 0
nvme_smart_log{device="nvme0n1",metric="thm_temp1_trans_count"} 0
nvme_smart_log{device="nvme0n1",metric="thm_temp2_total_time"} 0
nvme_smart_log{device="nvme0n1",metric="thm_temp2_trans_count"} 0
nvme_smart_log{device="nvme0n1",metric="unsafe_shutdowns"} 139
nvme_smart_log{device="nvme0n1",metric="warning_temp_time"} 0
`
	if err := testutil.GatherAndCompare(registry, strings.NewReader(expectedMetrics), "nvme_smart_log"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}

	// Simulate the drive becoming absent
	getNVMeDrives = mockGetNVMeDrivesAbsent

	go metrics.UpdateMetrics()

	// Allow some time for metrics to be updated
	time.Sleep(1 * time.Second)

	// Test that the metrics for the absent drive are still present
	if err := testutil.GatherAndCompare(registry, strings.NewReader(expectedMetrics), "nvme_smart_log"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}
}

func TestUpdateMetricsWithDriveAbsence(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		MockOutput: `{
  "critical_warning" : 0,
  "temperature" : 301,
  "avail_spare" : 100,
  "spare_thresh" : 5,
  "percent_used" : 15,
  "data_units_read" : 499296134,
  "data_units_written" : 1474968593,
  "host_read_commands" : 9347931143,
  "host_write_commands" : 51493840602,
  "controller_busy_time" : 1974546,
  "power_cycles" : 290,
  "power_on_hours" : 38313,
  "unsafe_shutdowns" : 139,
  "media_errors" : 0,
  "num_err_log_entries" : 17,
  "warning_temp_time" : 0,
  "critical_comp_time" : 0,
  "temperature_sensor_1" : 306,
  "temperature_sensor_2" : 301,
  "temperature_sensor_3" : 296,
  "temperature_sensor_4" : 295,
  "thm_temp1_trans_count" : 0,
  "thm_temp2_trans_count" : 0,
  "thm_temp1_total_time" : 0,
  "thm_temp2_total_time" : 0
}`,
	}

	// Mock getNVMeDrives function by setting it to the mock function
	originalGetNVMeDrives := getNVMeDrives
	getNVMeDrives = mockGetNVMeDrives
	defer func() { getNVMeDrives = originalGetNVMeDrives }()

	registry := prometheus.NewRegistry()
	metrics := NewMetrics(mockExecutor, registry)

	go metrics.UpdateMetrics()

	// Allow some time for metrics to be updated
	time.Sleep(1 * time.Second)

	// Test SMART log metrics
	expectedMetrics := `
# HELP nvme_smart_log SMART log metrics for NVMe devices
# TYPE nvme_smart_log gauge
nvme_smart_log{device="nvme0n1",metric="avail_spare"} 100
nvme_smart_log{device="nvme0n1",metric="controller_busy_time"} 1.974546e+06
nvme_smart_log{device="nvme0n1",metric="critical_comp_time"} 0
nvme_smart_log{device="nvme0n1",metric="critical_warning"} 0
nvme_smart_log{device="nvme0n1",metric="data_units_read"} 4.99296134e+08
nvme_smart_log{device="nvme0n1",metric="data_units_written"} 1.474968593e+09
nvme_smart_log{device="nvme0n1",metric="host_read_commands"} 9.347931143e+09
nvme_smart_log{device="nvme0n1",metric="host_write_commands"} 5.1493840602e+10
nvme_smart_log{device="nvme0n1",metric="media_errors"} 0
nvme_smart_log{device="nvme0n1",metric="num_err_log_entries"} 17
nvme_smart_log{device="nvme0n1",metric="percent_used"} 15
nvme_smart_log{device="nvme0n1",metric="power_cycles"} 290
nvme_smart_log{device="nvme0n1",metric="power_on_hours"} 38313
nvme_smart_log{device="nvme0n1",metric="spare_thresh"} 5
nvme_smart_log{device="nvme0n1",metric="temperature"} 301
nvme_smart_log{device="nvme0n1",metric="temperature_sensor_1"} 306
nvme_smart_log{device="nvme0n1",metric="temperature_sensor_2"} 301
nvme_smart_log{device="nvme0n1",metric="temperature_sensor_3"} 296
nvme_smart_log{device="nvme0n1",metric="temperature_sensor_4"} 295
nvme_smart_log{device="nvme0n1",metric="thm_temp1_total_time"} 0
nvme_smart_log{device="nvme0n1",metric="thm_temp1_trans_count"} 0
nvme_smart_log{device="nvme0n1",metric="thm_temp2_total_time"} 0
nvme_smart_log{device="nvme0n1",metric="thm_temp2_trans_count"} 0
nvme_smart_log{device="nvme0n1",metric="unsafe_shutdowns"} 139
nvme_smart_log{device="nvme0n1",metric="warning_temp_time"} 0
`
	if err := testutil.GatherAndCompare(registry, strings.NewReader(expectedMetrics), "nvme_smart_log"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}

	// Simulate the drive becoming absent
	getNVMeDrives = mockGetNVMeDrivesAbsent

	go metrics.UpdateMetrics()

	// Allow some time for metrics to be updated
	time.Sleep(1 * time.Second)

	// Test that the metrics for the absent drive are still present
	if err := testutil.GatherAndCompare(registry, strings.NewReader(expectedMetrics), "nvme_smart_log"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}
}
