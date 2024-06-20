package idrac

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

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

func TestGetRAIDStatus(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
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

	registry := prometheus.NewRegistry()
	client := NewClient(mockExecutor, registry)
	status, err := client.GetRAIDStatus()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(status) != 2 {
		t.Fatalf("Expected 2 RAID statuses, got %d", len(status))
	}
	if status["RAID.Integrated.1-1"]["Layout"] != "Raid-10" {
		t.Fatalf("Expected Layout to be Raid-10, got %s", status["RAID.Integrated.1-1"]["Layout"])
	}
	if status["RAID.Integrated.1-1"]["Size"] != "1787.50 GB" {
		t.Fatalf("Expected Size to be 1787.50 GB, got %s", status["RAID.Integrated.1-1"]["Size"])
	}
	if status["RAID.Integrated.1-0"]["Layout"] != "Raid-1" {
		t.Fatalf("Expected Layout to be Raid-1, got %s", status["RAID.Integrated.1-0"]["Layout"])
	}
	if status["RAID.Integrated.1-0"]["Size"] != "372.00 GB" {
		t.Fatalf("Expected Size to be 372.00 GB, got %s", status["RAID.Integrated.1-0"]["Size"])
	}
}

func TestGetRAIDStatusError(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		MockError: errors.New("command error"),
	}

	registry := prometheus.NewRegistry()
	client := NewClient(mockExecutor, registry)
	_, err := client.GetRAIDStatus()
	if err == nil {
		t.Fatalf("Expected error, got none")
	}
}

func TestUpdateMetrics(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
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

	registry := prometheus.NewRegistry()
	client := NewClient(mockExecutor, registry)

	go client.UpdateMetrics()

	// Allow some time for metrics to be updated
	time.Sleep(1 * time.Second)

	// Test RAID status metrics
	expectedStatus := `
# HELP raid_status Status of the RAID controller
# TYPE raid_status gauge
raid_status{vdisk="RAID.Integrated.1-1"} 1
raid_status{vdisk="RAID.Integrated.1-0"} 1
`
	if err := testutil.GatherAndCompare(registry, strings.NewReader(expectedStatus), "raid_status"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}

	// Test RAID redundancy metrics
	expectedRedundancy := `
# HELP raid_redundancy Remaining redundancy of the RAID controller
# TYPE raid_redundancy gauge
raid_redundancy{vdisk="RAID.Integrated.1-1"} 1
raid_redundancy{vdisk="RAID.Integrated.1-0"} 1
`
	if err := testutil.GatherAndCompare(registry, strings.NewReader(expectedRedundancy), "raid_redundancy"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}

	// Test RAID size metrics
	expectedSize := `
# HELP raid_size Size of the RAID controller
# TYPE raid_size gauge
raid_size{vdisk="RAID.Integrated.1-1"} 1787.5
raid_size{vdisk="RAID.Integrated.1-0"} 372
`
	if err := testutil.GatherAndCompare(registry, strings.NewReader(expectedSize), "raid_size"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}

	// Test RAID layout metrics
	expectedLayout := `
# HELP raid_layout Layout of the RAID controller
# TYPE raid_layout gauge
raid_layout{vdisk="RAID.Integrated.1-1"} 1
raid_layout{vdisk="RAID.Integrated.1-0"} 1
`
	if err := testutil.GatherAndCompare(registry, strings.NewReader(expectedLayout), "raid_layout"); err != nil {
		t.Fatalf("unexpected collecting result:\n%s", err)
	}
}
