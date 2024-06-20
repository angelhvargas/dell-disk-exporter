package main

import (
	"strings"
	"testing"
	"time"

	"github.com/angelhvargas/dell-disk-exporter/pkg/idrac"
	"github.com/angelhvargas/dell-disk-exporter/pkg/smart"

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

func TestCombinedMetrics(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		MockOutput: `
{
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

	mockRAIDExecutor := &MockCommandExecutor{
		MockOutput: `
Disk.Virtual.1:RAID.Integrated.1-1
   Layout                           = Raid-10
   Status                           = Ok
   RemainingRedundancy              = 1
   Size                             = 1787.50 GB
Disk.Virtual.0:RAID.Integrated.1-0
   Layout                           = Raid-1
   Status                           = Ok
   RemainingRedundancy              = 1
   Size                             = 372.00 GB
`,
	}

	// Set up SMART metrics
	smartRegistry := prometheus.NewRegistry()
	smartMetrics := smart.NewMetrics(mockExecutor, smartRegistry, 5*time.Minute)
	originalGetNVMeDrives := smart.GetNVMeDrives
	smart.GetNVMeDrives = mockGetNVMeDrives
	defer func() { smart.GetNVMeDrives = originalGetNVMeDrives }()

	go smartMetrics.UpdateMetrics()

	// Set up RAID metrics
	raidRegistry := prometheus.NewRegistry()
	raidClient := idrac.NewClient(mockRAIDExecutor, raidRegistry)
	go raidClient.UpdateMetrics()

	// Allow some time for metrics to be updated
	time.Sleep(1 * time.Second)

	// Test SMART log metrics
	expectedSmartMetrics := `
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
	if err := testutil.GatherAndCompare(smartRegistry, strings.NewReader(expectedSmartMetrics), "nvme_smart_log"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}

	// Test RAID status metrics
	expectedRaidStatus := `
# HELP raid_status Status of the RAID controller
# TYPE raid_status gauge
raid_status{vdisk="RAID.Integrated.1-1"} 1
raid_status{vdisk="RAID.Integrated.1-0"} 1
`
	if err := testutil.GatherAndCompare(raidRegistry, strings.NewReader(expectedRaidStatus), "raid_status"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}

	// Test RAID redundancy metrics
	expectedRaidRedundancy := `
# HELP raid_redundancy Remaining redundancy of the RAID controller
# TYPE raid_redundancy gauge
raid_redundancy{vdisk="RAID.Integrated.1-1"} 1
raid_redundancy{vdisk="RAID.Integrated.1-0"} 1
`
	if err := testutil.GatherAndCompare(raidRegistry, strings.NewReader(expectedRaidRedundancy), "raid_redundancy"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}

	// Test RAID size metrics
	expectedRaidSize := `
# HELP raid_size Size of the RAID controller
# TYPE raid_size gauge
raid_size{vdisk="RAID.Integrated.1-1"} 1787.5
raid_size{vdisk="RAID.Integrated.1-0"} 372
`
	if err := testutil.GatherAndCompare(raidRegistry, strings.NewReader(expectedRaidSize), "raid_size"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}

	// Test RAID layout metrics
	expectedRaidLayout := `
# HELP raid_layout Layout of the RAID controller
# TYPE raid_layout gauge
raid_layout{vdisk="RAID.Integrated.1-1"} 1
raid_layout{vdisk="RAID.Integrated.1-0"} 1
`
	if err := testutil.GatherAndCompare(raidRegistry, strings.NewReader(expectedRaidLayout), "raid_layout"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}
}
