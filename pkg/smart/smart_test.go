package smart

import (
	"testing"
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

	metrics := NewMetrics(mockExecutor)
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
