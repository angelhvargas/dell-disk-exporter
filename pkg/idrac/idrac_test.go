package idrac

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

	client := NewClient(mockExecutor)
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
